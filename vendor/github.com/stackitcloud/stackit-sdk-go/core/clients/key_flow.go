package clients

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// Service Account Key Flow
	// Auth flow env variables
	ServiceAccountKey     = "STACKIT_SERVICE_ACCOUNT_KEY"
	PrivateKey            = "STACKIT_PRIVATE_KEY"
	ServiceAccountKeyPath = "STACKIT_SERVICE_ACCOUNT_KEY_PATH"
	PrivateKeyPath        = "STACKIT_PRIVATE_KEY_PATH"
	tokenAPI              = "https://service-account.api.stackit.cloud/token" //nolint:gosec // linter false positive
	defaultTokenType      = "Bearer"
	defaultScope          = ""
)

var _ AuthFlow = &KeyFlow{}

// KeyFlow handles auth with SA key
type KeyFlow struct {
	rt            http.RoundTripper
	authClient    *http.Client
	config        *KeyFlowConfig
	key           *ServiceAccountKeyResponse
	privateKey    *rsa.PrivateKey
	privateKeyPEM []byte

	tokenMutex sync.RWMutex
	token      *TokenResponseBody

	// If the current access token would expire in less than TokenExpirationLeeway,
	// the client will refresh it early to prevent clock skew or other timing issues.
	tokenExpirationLeeway time.Duration
}

// KeyFlowConfig is the flow config
type KeyFlowConfig struct {
	ServiceAccountKey *ServiceAccountKeyResponse
	PrivateKey        string
	// Deprecated: retry options were removed to reduce complexity of the client. If this functionality is needed, you can provide your own custom HTTP client.
	ClientRetry                   *RetryConfig
	TokenUrl                      string
	BackgroundTokenRefreshContext context.Context // Functionality is enabled if this isn't nil
	HTTPTransport                 http.RoundTripper
	AuthHTTPClient                *http.Client
}

// ServiceAccountKeyResponse is the API response
// when creating a new SA key
type ServiceAccountKeyResponse struct {
	Active       bool                          `json:"active"`
	CreatedAt    time.Time                     `json:"createdAt"`
	Credentials  *ServiceAccountKeyCredentials `json:"credentials"`
	ID           uuid.UUID                     `json:"id"`
	KeyAlgorithm string                        `json:"keyAlgorithm"`
	KeyOrigin    string                        `json:"keyOrigin"`
	KeyType      string                        `json:"keyType"`
	PublicKey    string                        `json:"publicKey"`
	ValidUntil   *time.Time                    `json:"validUntil,omitempty"`
}

type ServiceAccountKeyCredentials struct {
	Aud        string    `json:"aud"`
	Iss        string    `json:"iss"`
	Kid        string    `json:"kid"`
	PrivateKey *string   `json:"privateKey,omitempty"`
	Sub        uuid.UUID `json:"sub"`
}

// GetConfig returns the flow configuration
func (c *KeyFlow) GetConfig() KeyFlowConfig {
	if c.config == nil {
		return KeyFlowConfig{}
	}
	return *c.config
}

// GetServiceAccountEmail returns the service account email
func (c *KeyFlow) GetServiceAccountEmail() string {
	if c.key == nil {
		return ""
	}
	return c.key.Credentials.Iss
}

// GetToken returns the token field
//
// Deprecated: Use GetAccessToken instead.
// This will be removed after 2026-07-01.
func (c *KeyFlow) GetToken() TokenResponseBody {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()

	if c.token == nil {
		return TokenResponseBody{}
	}
	// Returned struct is passed by value (because it's a struct)
	// So no deepy copy needed
	return *c.token
}

func (c *KeyFlow) Init(cfg *KeyFlowConfig) error {
	// No concurrency at this point, so no mutex check needed
	c.token = &TokenResponseBody{}
	c.config = cfg

	if c.config.TokenUrl == "" {
		c.config.TokenUrl = tokenAPI
	}

	c.tokenExpirationLeeway = defaultTokenExpirationLeeway

	if c.rt = cfg.HTTPTransport; c.rt == nil {
		c.rt = http.DefaultTransport
	}

	if c.authClient = cfg.AuthHTTPClient; cfg.AuthHTTPClient == nil {
		c.authClient = &http.Client{
			Transport: c.rt,
			Timeout:   DefaultClientTimeout,
		}
	}

	err := c.validate()
	if err != nil {
		return err
	}
	if c.config.BackgroundTokenRefreshContext != nil {
		go continuousRefreshToken(c)
	}
	return nil
}

// SetToken can be used to set an access and refresh token manually in the client.
// The other fields in the token field are determined by inspecting the token or setting default values.
//
// Deprecated: This method will be removed in future versions. Access tokens are going to be automatically managed by the client.
// This will be removed after 2026-07-01.
func (c *KeyFlow) SetToken(accessToken, refreshToken string) error {
	// We can safely use ParseUnverified because we are not authenticating the user,
	// We are parsing the token just to get the expiration time claim
	parsedAccessToken, _, err := jwt.NewParser().ParseUnverified(accessToken, &jwt.RegisteredClaims{})
	if err != nil {
		return fmt.Errorf("parse access token to read expiration time: %w", err)
	}
	exp, err := parsedAccessToken.Claims.GetExpirationTime()
	if err != nil {
		return fmt.Errorf("get expiration time from access token: %w", err)
	}

	c.tokenMutex.Lock()
	c.token = &TokenResponseBody{
		AccessToken:  accessToken,
		ExpiresIn:    int(exp.Time.Unix()),
		RefreshToken: refreshToken,
		Scope:        defaultScope,
		TokenType:    defaultTokenType,
	}
	c.tokenMutex.Unlock()
	return nil
}

// Roundtrip performs the request
func (c *KeyFlow) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.rt == nil {
		return nil, fmt.Errorf("please run Init()")
	}

	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	return c.rt.RoundTrip(req)
}

// GetAccessToken returns a short-lived access token and saves the access and refresh tokens in the token field
func (c *KeyFlow) GetAccessToken() (string, error) {
	if c.rt == nil {
		return "", fmt.Errorf("nil http round tripper, please run Init()")
	}
	var accessToken string

	c.tokenMutex.RLock()
	if c.token != nil {
		accessToken = c.token.AccessToken
	}
	c.tokenMutex.RUnlock()

	accessTokenExpired, err := tokenExpired(accessToken, c.tokenExpirationLeeway)
	if err != nil {
		return "", fmt.Errorf("check access token is expired: %w", err)
	}
	if !accessTokenExpired {
		return accessToken, nil
	}
	if err = c.recreateAccessToken(); err != nil {
		var oapiErr *oapierror.GenericOpenAPIError
		if ok := errors.As(err, &oapiErr); ok {
			reg := regexp.MustCompile("Key with kid .*? was not found")
			if reg.Match(oapiErr.Body) {
				err = fmt.Errorf("check if your configured key is valid and if the token endpoint is configured correct: %w", err)
			}
		}
		return "", fmt.Errorf("get new access token: %w", err)
	}

	c.tokenMutex.RLock()
	accessToken = c.token.AccessToken
	c.tokenMutex.RUnlock()

	return accessToken, nil
}

func (c *KeyFlow) refreshAccessToken() error {
	return c.recreateAccessToken()
}

func (c *KeyFlow) getBackgroundTokenRefreshContext() context.Context {
	return c.config.BackgroundTokenRefreshContext
}

// validate the client is configured well
func (c *KeyFlow) validate() error {
	if c.config.ServiceAccountKey == nil {
		return fmt.Errorf("service account access key cannot be empty")
	}
	if c.config.PrivateKey == "" {
		return fmt.Errorf("private key cannot be empty")
	}

	c.key = c.config.ServiceAccountKey
	var err error
	c.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(c.config.PrivateKey))
	if err != nil {
		return fmt.Errorf("parse private key from PEM file: %w", err)
	}

	// Encode the private key in PEM format
	privKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(c.privateKey),
	}
	c.privateKeyPEM = pem.EncodeToMemory(privKeyPEM)

	if c.tokenExpirationLeeway < 0 {
		return fmt.Errorf("token expiration leeway cannot be negative")
	}

	return nil
}

// Flow auth functions

// recreateAccessToken is used to create a new access token
// when the existing one isn't valid anymore
func (c *KeyFlow) recreateAccessToken() error {
	var refreshToken string

	c.tokenMutex.RLock()
	if c.token != nil {
		refreshToken = c.token.RefreshToken
	}
	c.tokenMutex.RUnlock()

	refreshTokenExpired, err := tokenExpired(refreshToken, c.tokenExpirationLeeway)
	if err != nil {
		return err
	}
	if !refreshTokenExpired {
		return c.createAccessTokenWithRefreshToken()
	}
	return c.createAccessToken()
}

// createAccessToken creates an access token using self signed JWT
func (c *KeyFlow) createAccessToken() (err error) {
	grant := "urn:ietf:params:oauth:grant-type:jwt-bearer"
	assertion, err := c.generateSelfSignedJWT()
	if err != nil {
		return err
	}
	res, err := c.requestToken(grant, assertion)
	if err != nil {
		return err
	}
	defer func() {
		tempErr := res.Body.Close()
		if tempErr != nil && err == nil {
			err = fmt.Errorf("close request access token response: %w", tempErr)
		}
	}()
	token, err := parseTokenResponse(res)
	if err != nil {
		return err
	}
	c.tokenMutex.Lock()
	c.token = token
	c.tokenMutex.Unlock()
	return nil
}

// createAccessTokenWithRefreshToken creates an access token using
// an existing pre-validated refresh token
// Deprecated: This method will be removed in future versions. Access tokens are going to be refreshed without refresh token.
// This will be removed after 2026-07-01.
func (c *KeyFlow) createAccessTokenWithRefreshToken() (err error) {
	c.tokenMutex.RLock()
	refreshToken := c.token.RefreshToken
	c.tokenMutex.RUnlock()

	res, err := c.requestToken("refresh_token", refreshToken)
	if err != nil {
		return err
	}
	defer func() {
		tempErr := res.Body.Close()
		if tempErr != nil && err == nil {
			err = fmt.Errorf("close request access token with refresh token response: %w", tempErr)
		}
	}()
	token, err := parseTokenResponse(res)
	if err != nil {
		return err
	}
	c.tokenMutex.Lock()
	c.token = token
	c.tokenMutex.Unlock()
	return nil
}

// generateSelfSignedJWT generates JWT token
func (c *KeyFlow) generateSelfSignedJWT() (string, error) {
	claims := jwt.MapClaims{
		"iss": c.key.Credentials.Iss,
		"sub": c.key.Credentials.Sub,
		"jti": uuid.New(),
		"aud": c.key.Credentials.Aud,
		"iat": jwt.NewNumericDate(time.Now()),
		"exp": jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	token.Header["kid"] = c.key.Credentials.Kid
	tokenString, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// requestToken makes a request to the SA token API
func (c *KeyFlow) requestToken(grant, assertion string) (*http.Response, error) {
	body := url.Values{}
	body.Set("grant_type", grant)
	if grant == "refresh_token" {
		body.Set("refresh_token", assertion)
	} else {
		body.Set("assertion", assertion)
	}

	payload := strings.NewReader(body.Encode())
	req, err := http.NewRequest(http.MethodPost, c.config.TokenUrl, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return c.authClient.Do(req)
}

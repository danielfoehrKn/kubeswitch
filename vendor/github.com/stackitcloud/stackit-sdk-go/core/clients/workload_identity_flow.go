package clients

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/stackitcloud/stackit-sdk-go/core/oidcadapters"
	"github.com/stackitcloud/stackit-sdk-go/core/utils"
)

const (
	clientIDEnv           = "STACKIT_SERVICE_ACCOUNT_EMAIL"
	FederatedTokenFileEnv = "STACKIT_FEDERATED_TOKEN_FILE"         //nolint:gosec // This is not a secret, just the env variable name
	wifTokenEndpointEnv   = "STACKIT_IDP_TOKEN_ENDPOINT"           //nolint:gosec // This is not a secret, just the env variable name
	wifTokenExpirationEnv = "STACKIT_IDP_TOKEN_EXPIRATION_SECONDS" //nolint:gosec // This is not a secret, just the env variable name

	wifClientAssertionType    = "urn:schwarz:params:oauth:client-assertion-type:workload-jwt"
	wifGrantType              = "client_credentials"
	defaultWifTokenEndpoint   = "https://accounts.stackit.cloud/oauth/v2/token"       //nolint:gosec // This is not a secret, just the public endpoint for default value
	defaultFederatedTokenPath = "/var/run/secrets/stackit.cloud/serviceaccount/token" //nolint:gosec // This is not a secret, just the default path for workload identity token
	defaultWifExpirationToken = "1h"
)

var (
	_ = utils.GetEnvOrDefault(wifTokenExpirationEnv, defaultWifExpirationToken) // Not used yet
)

var _ AuthFlow = &WorkloadIdentityFederationFlow{}

// WorkloadIdentityFlow handles auth with Workload Identity Federation
type WorkloadIdentityFederationFlow struct {
	rt         http.RoundTripper
	authClient *http.Client
	config     *WorkloadIdentityFederationFlowConfig

	tokenMutex sync.RWMutex
	token      *TokenResponseBody

	// If the current access token would expire in less than TokenExpirationLeeway,
	// the client will refresh it early to prevent clock skew or other timing issues.
	tokenExpirationLeeway time.Duration
}

// KeyFlowConfig is the flow config
type WorkloadIdentityFederationFlowConfig struct {
	TokenUrl                      string
	ClientID                      string
	TokenExpiration               string          // Not supported yet
	BackgroundTokenRefreshContext context.Context // Functionality is enabled if this isn't nil
	HTTPTransport                 http.RoundTripper
	AuthHTTPClient                *http.Client
	FederatedTokenFunction        oidcadapters.OIDCTokenFunc // Function to get the federated token
}

// GetConfig returns the flow configuration
func (c *WorkloadIdentityFederationFlow) GetConfig() WorkloadIdentityFederationFlowConfig {
	if c.config == nil {
		return WorkloadIdentityFederationFlowConfig{}
	}
	return *c.config
}

// GetAccessToken implements AuthFlow.
func (c *WorkloadIdentityFederationFlow) GetAccessToken() (string, error) {
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
	if err = c.createAccessToken(); err != nil {
		return "", fmt.Errorf("get new access token: %w", err)
	}

	c.tokenMutex.RLock()
	accessToken = c.token.AccessToken
	c.tokenMutex.RUnlock()

	return accessToken, nil
}

func (c *WorkloadIdentityFederationFlow) refreshAccessToken() error {
	return c.createAccessToken()
}

// RoundTrip implements the http.RoundTripper interface.
// It gets a token, adds it to the request's authorization header, and performs the request.
func (c *WorkloadIdentityFederationFlow) RoundTrip(req *http.Request) (*http.Response, error) {
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

// getBackgroundTokenRefreshContext implements AuthFlow.
func (c *WorkloadIdentityFederationFlow) getBackgroundTokenRefreshContext() context.Context {
	return c.config.BackgroundTokenRefreshContext
}

func (c *WorkloadIdentityFederationFlow) Init(cfg *WorkloadIdentityFederationFlowConfig) error {
	// No concurrency at this point, so no mutex check needed
	c.token = &TokenResponseBody{}
	c.config = cfg

	if c.config.TokenUrl == "" {
		c.config.TokenUrl = utils.GetEnvOrDefault(wifTokenEndpointEnv, defaultWifTokenEndpoint)
	}

	if c.config.ClientID == "" {
		c.config.ClientID = utils.GetEnvOrDefault(clientIDEnv, "")
	}

	if c.config.FederatedTokenFunction == nil {
		c.config.FederatedTokenFunction = oidcadapters.ReadJWTFromFileSystem(utils.GetEnvOrDefault(FederatedTokenFileEnv, defaultFederatedTokenPath))
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

// validate the client is configured well
func (c *WorkloadIdentityFederationFlow) validate() error {
	if c.config.ClientID == "" {
		return fmt.Errorf("client ID cannot be empty")
	}
	if c.config.TokenUrl == "" {
		return fmt.Errorf("token URL cannot be empty")
	}
	if _, err := c.config.FederatedTokenFunction(context.Background()); err != nil {
		return fmt.Errorf("error reading federated token file - %w", err)
	}
	if c.tokenExpirationLeeway < 0 {
		return fmt.Errorf("token expiration leeway cannot be negative")
	}

	return nil
}

// createAccessToken creates an access token using self signed JWT
func (c *WorkloadIdentityFederationFlow) createAccessToken() error {
	clientAssertion, err := c.config.FederatedTokenFunction(context.Background())
	if err != nil {
		return err
	}
	res, err := c.requestToken(c.config.ClientID, clientAssertion)
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

func (c *WorkloadIdentityFederationFlow) requestToken(clientID, assertion string) (*http.Response, error) {
	body := url.Values{}
	body.Set("grant_type", wifGrantType)
	body.Set("client_assertion_type", wifClientAssertionType)
	body.Set("client_assertion", assertion)
	body.Set("client_id", clientID)

	payload := strings.NewReader(body.Encode())
	req, err := http.NewRequest(http.MethodPost, c.config.TokenUrl, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return c.authClient.Do(req)
}

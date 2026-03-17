package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"
)

const (
	defaultTokenExpirationLeeway = time.Second * 5
)

type AuthFlow interface {
	RoundTrip(req *http.Request) (*http.Response, error)
	GetAccessToken() (string, error)
	getBackgroundTokenRefreshContext() context.Context
	refreshAccessToken() error
}

// TokenResponseBody is the API response
// when requesting a new token
type TokenResponseBody struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	// Deprecated: RefreshToken is no longer used and the SDK will not attempt to refresh tokens using it but will instead use the AuthFlow implementation to get new tokens.
	// This will be removed after 2026-07-01.
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

func parseTokenResponse(res *http.Response) (*TokenResponseBody, error) {
	if res == nil {
		return nil, fmt.Errorf("received bad response from API")
	}
	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			// Fail silently, omit body from error
			// We're trying to show error details, so it's unnecessary to fail because of this err
			body = []byte{}
		}
		return nil, &oapierror.GenericOpenAPIError{
			StatusCode: res.StatusCode,
			Body:       body,
		}
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	token := &TokenResponseBody{}
	err = json.Unmarshal(body, token)
	if err != nil {
		return nil, fmt.Errorf("unmarshal token response: %w", err)
	}
	return token, nil
}

func tokenExpired(token string, tokenExpirationLeeway time.Duration) (bool, error) {
	if token == "" {
		return true, nil
	}

	// We can safely use ParseUnverified because we are not authenticating the user at this point.
	// We're just checking the expiration time
	tokenParsed, _, err := jwt.NewParser().ParseUnverified(token, &jwt.RegisteredClaims{})
	if err != nil {
		return false, fmt.Errorf("parse token: %w", err)
	}

	expirationTimestampNumeric, err := tokenParsed.Claims.GetExpirationTime()
	if err != nil {
		return false, fmt.Errorf("get expiration timestamp: %w", err)
	}

	// Pretend to be `tokenExpirationLeeway` into the future to avoid token expiring
	// between retrieving the token and upstream systems validating it.
	now := time.Now().Add(tokenExpirationLeeway)
	return now.After(expirationTimestampNumeric.Time), nil
}

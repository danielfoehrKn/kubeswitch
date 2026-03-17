package clients

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"
)

var (
	defaultTimeStartBeforeTokenExpiration = 30 * time.Minute
	defaultTimeBetweenContextCheck        = time.Second
	defaultTimeBetweenTries               = 5 * time.Minute
)

// Continuously refreshes the token of an auth flow, retrying if the token API returns 5xx errrors. Writes to stderr when it terminates.
//
// To terminate this routine, close the context in flow.getBackgroundTokenRefreshContext().
func continuousRefreshToken(flow AuthFlow) {
	refresher := &continuousTokenRefresher{
		flow:                           flow,
		timeStartBeforeTokenExpiration: defaultTimeStartBeforeTokenExpiration,
		timeBetweenContextCheck:        defaultTimeBetweenContextCheck,
		timeBetweenTries:               defaultTimeBetweenTries,
	}
	err := refresher.continuousRefreshToken()
	fmt.Fprintf(os.Stderr, "Token refreshing terminated: %v", err)
}

type continuousTokenRefresher struct {
	flow AuthFlow
	// Token refresh tries start at [Access token expiration timestamp] - [This duration]
	timeStartBeforeTokenExpiration time.Duration
	timeBetweenContextCheck        time.Duration
	timeBetweenTries               time.Duration
}

// Continuously refreshes the token of an auth flow, retrying if the token API returns 5xx errrors. Always returns with a non-nil error.
//
// To terminate this routine, close the context in refresher.flow.getBackgroundTokenRefreshContext().
func (refresher *continuousTokenRefresher) continuousRefreshToken() error {
	// Compute timestamp where we'll refresh token
	// Access token may be empty at this point, we have to check it
	var startRefreshTimestamp time.Time

	expirationTimestamp, err := refresher.getAccessTokenExpirationTimestamp()
	if err != nil {
		return fmt.Errorf("get access token expiration timestamp: %w", err)
	}
	startRefreshTimestamp = expirationTimestamp.Add(-refresher.timeStartBeforeTokenExpiration)

	for {
		err := refresher.waitUntilTimestamp(startRefreshTimestamp)
		if err != nil {
			return err
		}

		err = refresher.flow.getBackgroundTokenRefreshContext().Err()
		if err != nil {
			return fmt.Errorf("check context: %w", err)
		}

		ok, err := refresher.refreshToken()
		if err != nil {
			return fmt.Errorf("refresh tokens: %w", err)
		}
		if !ok {
			startRefreshTimestamp = startRefreshTimestamp.Add(refresher.timeBetweenTries)
			continue
		}

		expirationTimestamp, err := refresher.getAccessTokenExpirationTimestamp()
		if err != nil {
			return fmt.Errorf("get access token expiration timestamp: %w", err)
		}
		startRefreshTimestamp = expirationTimestamp.Add(-refresher.timeStartBeforeTokenExpiration)
	}
}

func (refresher *continuousTokenRefresher) getAccessTokenExpirationTimestamp() (*time.Time, error) {
	accessToken, err := refresher.flow.GetAccessToken()
	if err != nil {
		return nil, err
	}

	// We can safely use ParseUnverified because we are not doing authentication of any kind
	// We're just checking the expiration time
	tokenParsed, _, err := jwt.NewParser().ParseUnverified(accessToken, &jwt.RegisteredClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	expirationTimestampNumeric, err := tokenParsed.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("get expiration timestamp: %w", err)
	}
	return &expirationTimestampNumeric.Time, nil
}

func (refresher *continuousTokenRefresher) waitUntilTimestamp(timestamp time.Time) error {
	for time.Now().Before(timestamp) {
		err := refresher.flow.getBackgroundTokenRefreshContext().Err()
		if err != nil {
			return fmt.Errorf("check context: %w", err)
		}
		time.Sleep(refresher.timeBetweenContextCheck)
	}
	return nil
}

// Returns:
//   - (true, nil) if successful.
//   - (false, nil) if not successful but should be retried.
//   - (_, err) if not successful and shouldn't be retried.
func (refresher *continuousTokenRefresher) refreshToken() (bool, error) {
	err := refresher.flow.refreshAccessToken()
	if err == nil {
		return true, nil
	}

	// Should be retried if this is an API error with status code 5xx
	oapiErr := &oapierror.GenericOpenAPIError{}
	if !errors.As(err, &oapiErr) {
		return false, err
	}
	if oapiErr.StatusCode < 500 {
		return false, err
	}
	return false, nil
}

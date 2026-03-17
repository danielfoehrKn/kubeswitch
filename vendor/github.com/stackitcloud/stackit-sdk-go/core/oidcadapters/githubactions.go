package oidcadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func RequestGHOIDCToken(oidcRequestUrl, oidcRequestToken string) OIDCTokenFunc {
	return func(ctx context.Context) (string, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, oidcRequestUrl, http.NoBody)
		if err != nil {
			return "", fmt.Errorf("githubAssertion: failed to build request: %w", err)
		}

		query, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			return "", fmt.Errorf("githubAssertion: cannot parse URL query")
		}

		if query.Get("audience") == "" {
			query.Set("audience", "sts.accounts.stackit.cloud")
			req.URL.RawQuery = query.Encode()
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oidcRequestToken))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("githubAssertion: cannot request token: %w", err)
		}

		defer func() {
			_ = resp.Body.Close()
		}()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return "", fmt.Errorf("githubAssertion: cannot parse response: %w", err)
		}

		if c := resp.StatusCode; c < 200 || c > 299 {
			return "", fmt.Errorf("githubAssertion: received HTTP status %d with response: %s", resp.StatusCode, body)
		}

		var tokenRes struct {
			Value string `json:"value"`
		}
		if err := json.Unmarshal(body, &tokenRes); err != nil {
			return "", fmt.Errorf("githubAssertion: cannot unmarshal response: %w", err)
		}

		return tokenRes.Value, nil
	}
}

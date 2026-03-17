package oidcadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	adoPipelineOIDCAPIVersion = "7.1"
	adoAudience               = "api://AzureADTokenExchange"
)

func RequestAzureDevOpsOIDCToken(oidcRequestUrl, oidcRequestToken, serviceConnectionID string) OIDCTokenFunc {
	return func(ctx context.Context) (string, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, oidcRequestUrl, http.NoBody)
		if err != nil {
			return "", fmt.Errorf("azureDevOpsAssertion: failed to build request: %w", err)
		}

		query, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			return "", fmt.Errorf("azureDevOpsAssertion: cannot parse URL query")
		}

		if query.Get("api-version") == "" {
			query.Add("api-version", adoPipelineOIDCAPIVersion)
		}

		if query.Get("serviceConnectionId") == "" && serviceConnectionID != "" {
			query.Add("serviceConnectionId", serviceConnectionID)
		}

		if query.Get("audience") == "" {
			query.Set("audience", adoAudience) // Azure DevOps requires this specific audience for OIDC tokens
		}

		req.URL.RawQuery = query.Encode()

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oidcRequestToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("azureDevOpsAssertion: cannot request token: %w", err)
		}

		defer func() {
			_ = resp.Body.Close()
		}()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return "", fmt.Errorf("azureDevOpsAssertion: cannot parse response: %w", err)
		}

		if c := resp.StatusCode; c < 200 || c > 299 {
			return "", fmt.Errorf("azureDevOpsAssertion: received HTTP status %d with response: %s", resp.StatusCode, body)
		}

		var tokenRes struct {
			Value *string `json:"oidcToken"`
		}
		if err := json.Unmarshal(body, &tokenRes); err != nil || tokenRes.Value == nil {
			return "", fmt.Errorf("azureDevOpsAssertion: cannot unmarshal response: %w", err)
		}

		return *tokenRes.Value, nil
	}
}

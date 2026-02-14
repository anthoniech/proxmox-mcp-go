package mcp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ProxmoxClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewProxmoxClient(baseURL, token string) *ProxmoxClient {
	return &ProxmoxClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // Proxmox commonly uses self-signed certificates
				},
			},
		},
	}
}

type apiResponse struct {
	Data json.RawMessage `json:"data"`
}

func (c *ProxmoxClient) do(ctx context.Context, method, path string, body io.Reader) (string, error) {
	u := c.BaseURL + "/api2/json" + path

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "PVEAPIToken="+c.Token)
	if body != nil && (method == http.MethodPost || method == http.MethodPut) {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	var pretty json.RawMessage
	if err := json.Unmarshal(apiResp.Data, &pretty); err != nil {
		return string(apiResp.Data), nil
	}
	formatted, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return string(apiResp.Data), nil
	}
	return string(formatted), nil
}

func (c *ProxmoxClient) Get(ctx context.Context, path string) (string, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *ProxmoxClient) Post(ctx context.Context, path string, data url.Values) (string, error) {
	var body io.Reader
	if data != nil {
		body = strings.NewReader(data.Encode())
	}
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *ProxmoxClient) Delete(ctx context.Context, path string, params url.Values) (string, error) {
	p := path
	if len(params) > 0 {
		p += "?" + params.Encode()
	}
	return c.do(ctx, http.MethodDelete, p, nil)
}

// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

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
	"time"

	log "github.com/sirupsen/logrus"
)

type ProxmoxClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	Logger     *log.Logger
}

func NewProxmoxClient(baseURL, token string, logger *log.Logger) *ProxmoxClient {
	return &ProxmoxClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		Logger:  logger,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // Proxmox commonly uses self-signed certificates
				},
			},
		},
	}
}

func (c *ProxmoxClient) logRequest(method, path string, statusCode, respBytes int, duration time.Duration, err error) {
	if c.Logger == nil {
		return
	}

	fields := log.Fields{
		"event.category": "api",
		"event.action":   "proxmox_api_call",
		"http.method":    method,
		"url.path":       path,
		"duration_ms":    duration.Milliseconds(),
	}

	if statusCode > 0 {
		fields["http.status"] = statusCode
	}
	if respBytes > 0 {
		fields["response.bytes"] = respBytes
	}

	entry := c.Logger.WithFields(fields)
	if err != nil {
		entry.WithError(err).Error("Proxmox API call failed")
	} else {
		entry.Info("Proxmox API call completed")
	}
}

type apiResponse struct {
	Data json.RawMessage `json:"data"`
}

func (c *ProxmoxClient) do(ctx context.Context, method, path string, body io.Reader) (string, error) {
	start := time.Now()
	u := c.BaseURL + "/api2/json" + path

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		reqErr := fmt.Errorf("creating request: %w", err)
		c.logRequest(method, path, 0, 0, time.Since(start), reqErr)
		return "", reqErr
	}

	req.Header.Set("Authorization", "PVEAPIToken="+c.Token)
	if body != nil && (method == http.MethodPost || method == http.MethodPut) {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		reqErr := fmt.Errorf("request failed: %w", err)
		c.logRequest(method, path, 0, 0, time.Since(start), reqErr)
		return "", reqErr
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		readErr := fmt.Errorf("reading response: %w", err)
		c.logRequest(method, path, resp.StatusCode, 0, time.Since(start), readErr)
		return "", readErr
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
		c.logRequest(method, path, resp.StatusCode, len(respBody), time.Since(start), apiErr)
		return "", apiErr
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		parseErr := fmt.Errorf("parsing response: %w", err)
		c.logRequest(method, path, resp.StatusCode, len(respBody), time.Since(start), parseErr)
		return "", parseErr
	}

	c.logRequest(method, path, resp.StatusCode, len(respBody), time.Since(start), nil)

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

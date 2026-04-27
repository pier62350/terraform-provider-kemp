// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

// Package loadmaster is a thin Go client for the Progress Kemp LoadMaster
// JSON RPC API (the "/accessv2" endpoint).
//
// All commands are POSTed as JSON to <baseURL>/accessv2 with the command
// name in the "cmd" field plus authentication credentials. Responses are
// decoded into per-command response types.
package loadmaster

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a Kemp LoadMaster API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	username   string
	password   string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient lets the caller supply a pre-configured *http.Client (for
// custom timeouts, proxies, transports, etc.).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithInsecureSkipVerify disables TLS certificate verification. LoadMasters
// frequently use self-signed certificates, so this is enabled by default and
// can be turned off with WithInsecureSkipVerify(false).
func WithInsecureSkipVerify(skip bool) Option {
	return func(c *Client) {
		tr, ok := c.httpClient.Transport.(*http.Transport)
		if !ok {
			tr = &http.Transport{}
		}
		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{}
		}
		tr.TLSClientConfig.InsecureSkipVerify = skip
		c.httpClient.Transport = tr
	}
}

// WithAPIKey configures API-key authentication.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithBasicAuth configures username/password authentication.
func WithBasicAuth(user, pass string) Option {
	return func(c *Client) {
		c.username = user
		c.password = pass
	}
}

// NewClient builds a Client targeting the given LoadMaster base URL
// (e.g. "https://192.168.1.155:9443"). Options override the defaults.
func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// call executes a single LoadMaster API command. params, if non-nil, must
// marshal to a JSON object whose keys will be merged with "cmd" and
// authentication. out receives the decoded response.
func (c *Client) call(ctx context.Context, cmd string, params any, out any) error {
	body := map[string]any{"cmd": cmd}

	switch {
	case c.apiKey != "":
		body["apikey"] = c.apiKey
	case c.username != "":
		body["apiuser"] = c.username
		body["apipass"] = c.password
	default:
		return fmt.Errorf("loadmaster: no credentials configured (set api_key or username/password)")
	}

	if params != nil {
		raw, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("loadmaster: marshal params: %w", err)
		}
		var extra map[string]any
		if err := json.Unmarshal(raw, &extra); err != nil {
			return fmt.Errorf("loadmaster: params is not a JSON object: %w", err)
		}
		for k, v := range extra {
			body[k] = v
		}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("loadmaster: marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/accessv2", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("loadmaster: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("loadmaster: request to %s failed: %w", req.URL.String(), err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("loadmaster: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to surface a structured error; fall back to raw body.
		apiErr := &Error{HTTPStatus: resp.StatusCode}
		if jerr := json.Unmarshal(respBody, apiErr); jerr != nil || apiErr.Message == "" {
			apiErr.Message = string(respBody)
		}
		return apiErr
	}

	// Even at HTTP 200, LoadMaster may return {"status":"fail",...}.
	var head Response
	if err := json.Unmarshal(respBody, &head); err == nil && head.Status == "fail" {
		return &Error{HTTPStatus: resp.StatusCode, Code: head.Code, Status: head.Status, Message: head.Message}
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("loadmaster: decode response: %w (body=%s)", err, string(respBody))
	}
	return nil
}

// Response is the common envelope returned by every /accessv2 command.
type Response struct {
	Code    int    `json:"code"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

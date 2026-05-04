// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"fmt"
	"strconv"
)

// L7Limit represents a client IP rate-limit entry returned by a *list command.
type L7Limit struct {
	Addr  string `json:"Addr"`
	Limit int64  `json:"Limit"`
}

// URLLimitRule represents a URL-based rate limiting rule.
type URLLimitRule struct {
	Name    string `json:"Name"`
	Pattern string `json:"Pattern"`
	Limit   int64  `json:"Limit"`
	Match   int64  `json:"Match"`
}

// l7LimitListResponse is the response envelope for *limitlist commands.
type l7LimitListResponse struct {
	Response
	Data []L7Limit `json:"Data"`
}

// urlLimitRuleListResponse is the response envelope for listlimitrules.
type urlLimitRuleListResponse struct {
	Response
	Data []URLLimitRule `json:"Data"`
}

// --- Bandwidth limit ---

func (c *Client) AddBandwidthLimit(ctx context.Context, addr, limit string) error {
	type body struct {
		Addr  string `json:"l7addr"`
		Limit string `json:"l7limit"`
	}
	return c.call(ctx, "clientbandwidthlimitadd", body{Addr: addr, Limit: limit}, nil)
}

func (c *Client) DeleteBandwidthLimit(ctx context.Context, addr string) error {
	type body struct {
		Addr string `json:"l7addr"`
	}
	return c.call(ctx, "clientbandwidthlimitdel", body{Addr: addr}, nil)
}

func (c *Client) ListBandwidthLimits(ctx context.Context) ([]L7Limit, error) {
	var resp l7LimitListResponse
	if err := c.call(ctx, "clientbandwidthlimitlist", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// --- CPS limit (connections per second) ---

func (c *Client) AddCPSLimit(ctx context.Context, addr, limit string) error {
	type body struct {
		Addr  string `json:"l7addr"`
		Limit string `json:"l7limit"`
	}
	return c.call(ctx, "clientcpslimitadd", body{Addr: addr, Limit: limit}, nil)
}

func (c *Client) DeleteCPSLimit(ctx context.Context, addr string) error {
	type body struct {
		Addr string `json:"l7addr"`
	}
	return c.call(ctx, "clientcpslimitdel", body{Addr: addr}, nil)
}

func (c *Client) ListCPSLimits(ctx context.Context) ([]L7Limit, error) {
	var resp l7LimitListResponse
	if err := c.call(ctx, "clientcpslimitlist", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// --- RPS limit (requests per second) ---

func (c *Client) AddRPSLimit(ctx context.Context, addr, limit string) error {
	type body struct {
		Addr  string `json:"l7addr"`
		Limit string `json:"l7limit"`
	}
	return c.call(ctx, "clientrpslimitadd", body{Addr: addr, Limit: limit}, nil)
}

func (c *Client) DeleteRPSLimit(ctx context.Context, addr string) error {
	type body struct {
		Addr string `json:"l7addr"`
	}
	return c.call(ctx, "clientrpslimitdel", body{Addr: addr}, nil)
}

func (c *Client) ListRPSLimits(ctx context.Context) ([]L7Limit, error) {
	var resp l7LimitListResponse
	if err := c.call(ctx, "clientrpslimitlist", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// --- Connection limit (max concurrent connections) ---

func (c *Client) AddConnectionLimit(ctx context.Context, addr, limit string) error {
	type body struct {
		Addr  string `json:"l7addr"`
		Limit string `json:"l7limit"`
	}
	return c.call(ctx, "clientmaxclimitadd", body{Addr: addr, Limit: limit}, nil)
}

func (c *Client) DeleteConnectionLimit(ctx context.Context, addr string) error {
	type body struct {
		Addr string `json:"l7addr"`
	}
	return c.call(ctx, "clientmaxclimitdel", body{Addr: addr}, nil)
}

func (c *Client) ListConnectionLimits(ctx context.Context) ([]L7Limit, error) {
	var resp l7LimitListResponse
	if err := c.call(ctx, "clientmaxclimitlist", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// --- URL-based limit rules ---

func (c *Client) AddURLLimitRule(ctx context.Context, name, pattern string, limit, match int64) error {
	type body struct {
		Name    string `json:"name"`
		Pattern string `json:"pattern"`
		Limit   string `json:"limit"`
		Match   string `json:"match"`
	}
	return c.call(ctx, "addlimitrule", body{
		Name:    name,
		Pattern: pattern,
		Limit:   strconv.FormatInt(limit, 10),
		Match:   strconv.FormatInt(match, 10),
	}, nil)
}

func (c *Client) ModifyURLLimitRule(ctx context.Context, name, pattern string, limit, match int64) error {
	type body struct {
		Name    string `json:"name"`
		Pattern string `json:"pattern"`
		Limit   string `json:"limit"`
		Match   string `json:"match"`
	}
	return c.call(ctx, "modlimitrule", body{
		Name:    name,
		Pattern: pattern,
		Limit:   strconv.FormatInt(limit, 10),
		Match:   strconv.FormatInt(match, 10),
	}, nil)
}

func (c *Client) DeleteURLLimitRule(ctx context.Context, name string) error {
	type body struct {
		Name string `json:"name"`
	}
	return c.call(ctx, "dellimitrule", body{Name: name}, nil)
}

func (c *Client) ListURLLimitRules(ctx context.Context) ([]URLLimitRule, error) {
	var resp urlLimitRuleListResponse
	if err := c.call(ctx, "listlimitrules", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// FindURLLimitRule returns the named rule or nil if not found.
func (c *Client) FindURLLimitRule(ctx context.Context, name string) (*URLLimitRule, error) {
	rules, err := c.ListURLLimitRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("listlimitrules: %w", err)
	}
	for i := range rules {
		if rules[i].Name == name {
			return &rules[i], nil
		}
	}
	return nil, nil
}

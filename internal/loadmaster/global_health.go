// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type GlobalHealth struct {
	RetryInterval string `json:"RetryInterval"`
	Timeout       string `json:"Timeout"`
	RetryCount    string `json:"RetryCount"`
}

type GlobalHealthParams struct {
	RetryInterval string `json:"RetryInterval,omitempty"`
	Timeout       string `json:"Timeout,omitempty"`
	RetryCount    string `json:"RetryCount,omitempty"`
}

type globalHealthResponse struct {
	Response
	GlobalHealth
}

func (c *Client) ShowGlobalHealth(ctx context.Context) (*GlobalHealth, error) {
	var resp globalHealthResponse
	if err := c.call(ctx, "showhealth", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.GlobalHealth, nil
}

func (c *Client) ModifyGlobalHealth(ctx context.Context, p GlobalHealthParams) error {
	return c.call(ctx, "modhealth", p, nil)
}

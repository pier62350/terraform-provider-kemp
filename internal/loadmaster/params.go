// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type paramGetResponse struct {
	Response
	Value string `json:"Value"`
}

// GetParam reads a single named parameter via the generic "get" command.
func (c *Client) GetParam(ctx context.Context, param string) (string, error) {
	type body struct {
		Param string `json:"param"`
	}
	var resp paramGetResponse
	if err := c.call(ctx, "get", body{Param: param}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// SetParam writes a single named parameter via the generic "set" command.
func (c *Client) SetParam(ctx context.Context, param, value string) error {
	type body struct {
		Param string `json:"param"`
		Value string `json:"value"`
	}
	return c.call(ctx, "set", body{Param: param, Value: value}, nil)
}

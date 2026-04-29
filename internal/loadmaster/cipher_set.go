// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type CipherSet struct {
	Name    string
	Ciphers string // colon-separated on the wire
}

type cipherSetResponse struct {
	Response
	CipherSet string `json:"cipherset"`
}

// GetCipherSet reads a cipher set by name. Works for both built-in sets
// (Default, BestPractices, etc.) and custom sets.
func (c *Client) GetCipherSet(ctx context.Context, name string) (*CipherSet, error) {
	type body struct {
		Name string `json:"name"`
	}
	var resp cipherSetResponse
	if err := c.call(ctx, "getcipherset", body{Name: name}, &resp); err != nil {
		return nil, err
	}
	return &CipherSet{Name: name, Ciphers: resp.CipherSet}, nil
}

// ModifyCipherSet creates or updates a custom cipher set.
// If name already exists it is overwritten; if not it is created.
func (c *Client) ModifyCipherSet(ctx context.Context, name, value string) error {
	type body struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	return c.call(ctx, "modifycipherset", body{Name: name, Value: value}, nil)
}

// DeleteCipherSet removes a custom cipher set by name.
// Returns an error if the cipher set is still assigned to any virtual service.
func (c *Client) DeleteCipherSet(ctx context.Context, name string) error {
	type body struct {
		Name string `json:"name"`
	}
	return c.call(ctx, "delcipherset", body{Name: name}, nil)
}

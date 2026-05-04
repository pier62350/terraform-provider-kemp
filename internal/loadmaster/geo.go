// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type geoEnabledResponse struct {
	Response
	IsEnabled any `json:"IsEnabled,omitempty"`
	Data      any `json:"data,omitempty"`
}

// EnableGEO activates the GEO feature on the LoadMaster.
func (c *Client) EnableGEO(ctx context.Context) error {
	return c.call(ctx, "enablegeo", nil, nil)
}

// DisableGEO deactivates the GEO feature on the LoadMaster.
func (c *Client) DisableGEO(ctx context.Context) error {
	return c.call(ctx, "disablegeo", nil, nil)
}

// IsGEOEnabled returns the current GEO enabled state.
func (c *Client) IsGEOEnabled(ctx context.Context) (bool, error) {
	var resp geoEnabledResponse
	if err := c.call(ctx, "isgeoenabled", nil, &resp); err != nil {
		return false, err
	}
	if resp.IsEnabled != nil {
		return parseBoolAny(resp.IsEnabled), nil
	}
	return parseBoolAny(resp.Data), nil
}

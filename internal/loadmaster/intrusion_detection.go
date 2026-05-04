// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"strconv"
)

// GetIntrusionDetectionLevel returns the current paranoia/IPS detection level (0–4).
func (c *Client) GetIntrusionDetectionLevel(ctx context.Context) (int64, error) {
	value, err := c.GetParam(ctx, "paranoia")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(value, 10, 64)
}

// SetIntrusionDetectionLevel sets the paranoia/IPS detection level (0–4).
func (c *Client) SetIntrusionDetectionLevel(ctx context.Context, level int64) error {
	return c.SetParam(ctx, "paranoia", strconv.FormatInt(level, 10))
}

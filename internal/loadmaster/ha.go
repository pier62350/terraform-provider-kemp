// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"strconv"
)

// HAConfig holds the HA configuration read from the LoadMaster.
type HAConfig struct {
	Mode           int64
	PartnerAddress string
	SharedAddress  string
}

// GetHAConfig reads the HA mode and HA-related interface 0 addresses.
func (c *Client) GetHAConfig(ctx context.Context) (*HAConfig, error) {
	modeStr, err := c.GetParam(ctx, "hamode")
	if err != nil {
		return nil, err
	}
	mode, _ := strconv.ParseInt(modeStr, 10, 64)

	iface, err := c.ShowInterface(ctx, "0")
	if err != nil {
		return nil, err
	}

	return &HAConfig{
		Mode:           mode,
		PartnerAddress: iface.PartnerIPAddress,
		SharedAddress:  iface.SharedIPAddress,
	}, nil
}

// SetHAMode sets the HA mode (0=non-HA, 1=primary, 2=secondary).
func (c *Client) SetHAMode(ctx context.Context, mode int64) error {
	return c.SetParam(ctx, "hamode", strconv.FormatInt(mode, 10))
}

// SetHAPartnerAddress sets the partner LoadMaster IP on interface 0.
func (c *Client) SetHAPartnerAddress(ctx context.Context, addr string) error {
	type body struct {
		Interface        string `json:"interface"`
		PartnerIPAddress string `json:"PartnerIPAddress"`
	}
	return c.call(ctx, "modiface", body{Interface: "0", PartnerIPAddress: addr}, nil)
}

// SetHASharedAddress sets the shared virtual IP on interface 0.
func (c *Client) SetHASharedAddress(ctx context.Context, addr string) error {
	type body struct {
		Interface       string `json:"interface"`
		SharedIPAddress string `json:"SharedIPAddress"`
	}
	return c.call(ctx, "modiface", body{Interface: "0", SharedIPAddress: addr}, nil)
}

// SetHASecret sets the partner communication shared secret.
func (c *Client) SetHASecret(ctx context.Context, secret string) error {
	type body struct {
		Secret string `json:"secret"`
	}
	return c.call(ctx, "setlmcommsecret", body{Secret: secret}, nil)
}

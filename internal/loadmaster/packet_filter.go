// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// PacketRoutingFilter holds the current state of the packet routing (ACL) filter.
type PacketRoutingFilter struct {
	Enabled            bool
	Drop               bool
	RestrictToInterface bool
	IncludeWUI         bool
}

type aclBoolResponse struct {
	Response
	// The API may return a string "0"/"1" or a proper boolean.
	IsEnabled  interface{} `json:"IsEnabled"`
	IsDrop     interface{} `json:"IsDrop"`
	IsIfBlock  interface{} `json:"IsIfBlock"`
	IsWuiBlock interface{} `json:"IsWuiBlock"`
	// Fallback: some firmware versions return the value in "data".
	Data interface{} `json:"data"`
}

func parseBoolField(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val == "1" || val == "true" || val == "True"
	}
	return false
}

// GetPacketRoutingFilter reads all four ACL control flags.
func (c *Client) GetPacketRoutingFilter(ctx context.Context) (*PacketRoutingFilter, error) {
	var (
		enabledResp  aclBoolResponse
		dropResp     aclBoolResponse
		ifblockResp  aclBoolResponse
		wuiblockResp aclBoolResponse
	)

	if err := c.call(ctx, "aclcontrol.isenabled", nil, &enabledResp); err != nil {
		return nil, err
	}
	if err := c.call(ctx, "aclcontrol.isdrop", nil, &dropResp); err != nil {
		return nil, err
	}
	if err := c.call(ctx, "aclcontrol.isifblock", nil, &ifblockResp); err != nil {
		return nil, err
	}
	if err := c.call(ctx, "aclcontrol.iswuiblock", nil, &wuiblockResp); err != nil {
		return nil, err
	}

	return &PacketRoutingFilter{
		Enabled:             parseBoolField(pickBool(enabledResp.IsEnabled, enabledResp.Data)),
		Drop:                parseBoolField(pickBool(dropResp.IsDrop, dropResp.Data)),
		RestrictToInterface: parseBoolField(pickBool(ifblockResp.IsIfBlock, ifblockResp.Data)),
		IncludeWUI:          parseBoolField(pickBool(wuiblockResp.IsWuiBlock, wuiblockResp.Data)),
	}, nil
}

func pickBool(primary, fallback interface{}) interface{} {
	if primary != nil {
		return primary
	}
	return fallback
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// SetPacketRoutingFilter writes all four ACL control flags.
func (c *Client) SetPacketRoutingFilter(ctx context.Context, f PacketRoutingFilter) error {
	type enableBody struct {
		Enable string `json:"enable"`
	}
	type dropBody struct {
		Drop string `json:"drop"`
	}
	type ifblockBody struct {
		Ifblock string `json:"ifblock"`
	}
	type wuiblockBody struct {
		Wuiblock string `json:"wuiblock"`
	}

	if err := c.call(ctx, "aclcontrol", enableBody{Enable: boolToStr(f.Enabled)}, nil); err != nil {
		return err
	}
	if err := c.call(ctx, "aclcontrol", dropBody{Drop: boolToStr(f.Drop)}, nil); err != nil {
		return err
	}
	if err := c.call(ctx, "aclcontrol", ifblockBody{Ifblock: boolToStr(f.RestrictToInterface)}, nil); err != nil {
		return err
	}
	return c.call(ctx, "aclcontrol", wuiblockBody{Wuiblock: boolToStr(f.IncludeWUI)}, nil)
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// aclListResponse is the wire shape returned by aclcontrol list/listvs.
// LoadMaster returns the entries under an "Acl" key as a slice of strings.
type aclListResponse struct {
	Response
	Acl []string `json:"Acl,omitempty"`
}

// AddGlobalACLEntry adds an IP address or CIDR to the appliance-wide allow or
// block list. listType must be "allow" or "block".
func (c *Client) AddGlobalACLEntry(ctx context.Context, listType, addr string) error {
	type body struct {
		Add  string `json:"add"`
		Addr string `json:"addr"`
	}
	return c.call(ctx, "aclcontrol", body{Add: listType, Addr: addr}, nil)
}

// DeleteGlobalACLEntry removes an entry from the appliance-wide allow or block
// list. listType must be "allow" or "block".
func (c *Client) DeleteGlobalACLEntry(ctx context.Context, listType, addr string) error {
	type body struct {
		Del  string `json:"del"`
		Addr string `json:"addr"`
	}
	return c.call(ctx, "aclcontrol", body{Del: listType, Addr: addr}, nil)
}

// ListGlobalACLEntries returns all entries in the appliance-wide allow or
// block list. listType must be "allow" or "block".
func (c *Client) ListGlobalACLEntries(ctx context.Context, listType string) ([]string, error) {
	type body struct {
		List string `json:"list"`
	}
	var resp aclListResponse
	if err := c.call(ctx, "aclcontrol", body{List: listType}, &resp); err != nil {
		return nil, err
	}
	return resp.Acl, nil
}

// AddVSACLEntry adds an IP address to the per-VS allow or block list.
// vsID is the bare numeric VS Index. listType must be "allow" or "block".
func (c *Client) AddVSACLEntry(ctx context.Context, vsID, listType, addr string) error {
	type body struct {
		AddVS string `json:"addvs"`
		VSIP  string `json:"vsip"`
		Addr  string `json:"addr"`
	}
	return c.call(ctx, "aclcontrol", body{AddVS: listType, VSIP: vsID, Addr: addr}, nil)
}

// DeleteVSACLEntry removes an entry from the per-VS allow or block list.
// vsID is the bare numeric VS Index. listType must be "allow" or "block".
func (c *Client) DeleteVSACLEntry(ctx context.Context, vsID, listType, addr string) error {
	type body struct {
		DelVS string `json:"delvs"`
		VSIP  string `json:"vsip"`
		Addr  string `json:"addr"`
	}
	return c.call(ctx, "aclcontrol", body{DelVS: listType, VSIP: vsID, Addr: addr}, nil)
}

// ListVSACLEntries returns all entries in the per-VS allow or block list.
// vsID is the bare numeric VS Index. listType must be "allow" or "block".
func (c *Client) ListVSACLEntries(ctx context.Context, vsID, listType string) ([]string, error) {
	type body struct {
		ListVS string `json:"listvs"`
		VS     string `json:"vs"`
	}
	var resp aclListResponse
	if err := c.call(ctx, "aclcontrol", body{ListVS: listType, VS: vsID}, &resp); err != nil {
		return nil, err
	}
	return resp.Acl, nil
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// CreateSubVS attaches a new sub-virtual-service to the given parent VS.
//
// LoadMaster's API doesn't expose a dedicated "addsubvs" command — instead,
// you call modvs on the parent with the createsubvs trigger; the response
// is the freshly created SubVS, including its own Index. Subsequent CRUD
// on the SubVS uses the standard ShowVirtualService / ModifyVirtualService /
// DeleteVirtualService methods against the SubVS's Index.
func (c *Client) CreateSubVS(ctx context.Context, parentID string) (*VirtualService, error) {
	type body struct {
		VS          string `json:"vs"`
		CreateSubVS string `json:"createsubvs"`
	}
	var resp vsResponse
	if err := c.call(ctx, "modvs", body{VS: parentID, CreateSubVS: ""}, &resp); err != nil {
		return nil, err
	}
	return &resp.VirtualService, nil
}

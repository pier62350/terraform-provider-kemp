// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"fmt"
)

// ---- VLAN ----

// AddVLAN adds a VLAN to an interface.
func (c *Client) AddVLAN(ctx context.Context, interfaceID string, vlanID int64) error {
	type body struct {
		Interface string `json:"interface"`
		VlanID    string `json:"vlanid"`
	}
	return c.call(ctx, "addvlan", body{Interface: interfaceID, VlanID: fmt.Sprintf("%d", vlanID)}, nil)
}

// DeleteVLAN removes a VLAN from an interface.
func (c *Client) DeleteVLAN(ctx context.Context, interfaceID string, vlanID int64) error {
	type body struct {
		Interface string `json:"interface"`
		VlanID    string `json:"vlanid"`
	}
	return c.call(ctx, "delvlan", body{Interface: interfaceID, VlanID: fmt.Sprintf("%d", vlanID)}, nil)
}

// ---- VXLAN ----

// VXLANInfo represents a VXLAN tunnel as returned by showvxlan.
type VXLANInfo struct {
	Id     int32  `json:"Id"`
	Vni    int64  `json:"Vni"`
	Remote string `json:"Remote"`
}

type vxlanListResponse struct {
	Response
	Interface []VXLANInfo `json:"Interface,omitempty"`
}

// ShowVXLAN returns the VXLAN configuration for an interface.
func (c *Client) ShowVXLAN(ctx context.Context, interfaceID string) (*VXLANInfo, error) {
	type body struct {
		Interface string `json:"interface"`
	}
	var resp vxlanListResponse
	if err := c.call(ctx, "showvxlan", body{Interface: interfaceID}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Interface) == 0 {
		return nil, nil
	}
	return &resp.Interface[0], nil
}

// AddVXLAN creates a VXLAN tunnel on an interface.
func (c *Client) AddVXLAN(ctx context.Context, interfaceID string, vni int64, remote string) error {
	type body struct {
		Interface string `json:"interface"`
		Vni       string `json:"vni"`
		Remote    string `json:"remote"`
	}
	return c.call(ctx, "addvxlan", body{Interface: interfaceID, Vni: fmt.Sprintf("%d", vni), Remote: remote}, nil)
}

// ModifyVXLAN updates a VXLAN tunnel on an interface.
func (c *Client) ModifyVXLAN(ctx context.Context, interfaceID string, vni int64, remote string) error {
	type body struct {
		Interface string `json:"interface"`
		Vni       string `json:"vni"`
		Remote    string `json:"remote"`
	}
	return c.call(ctx, "modvxlan", body{Interface: interfaceID, Vni: fmt.Sprintf("%d", vni), Remote: remote}, nil)
}

// DeleteVXLAN removes a VXLAN tunnel from an interface.
func (c *Client) DeleteVXLAN(ctx context.Context, interfaceID string) error {
	type body struct {
		Interface string `json:"interface"`
	}
	return c.call(ctx, "delvxlan", body{Interface: interfaceID}, nil)
}

// ---- Additional Address ----

// AddInterfaceAddress adds an additional IP address to an interface.
func (c *Client) AddInterfaceAddress(ctx context.Context, interfaceID, addr string) error {
	type body struct {
		Interface string `json:"interface"`
		Addr      string `json:"addr"`
	}
	return c.call(ctx, "addaddress", body{Interface: interfaceID, Addr: addr}, nil)
}

// DeleteInterfaceAddress removes an additional IP address from an interface.
func (c *Client) DeleteInterfaceAddress(ctx context.Context, interfaceID, addr string) error {
	type body struct {
		Interface string `json:"interface"`
		Addr      string `json:"addr"`
	}
	return c.call(ctx, "deladdress", body{Interface: interfaceID, Addr: addr}, nil)
}

// ---- Bond ----

// CreateBond creates a bonded (LAG) interface.
func (c *Client) CreateBond(ctx context.Context, interfaceID string) error {
	type body struct {
		Interface string `json:"interface"`
	}
	return c.call(ctx, "createbond", body{Interface: interfaceID}, nil)
}

// UnbondInterface destroys a bonded interface.
func (c *Client) UnbondInterface(ctx context.Context, interfaceID string) error {
	type body struct {
		Interface string `json:"interface"`
	}
	return c.call(ctx, "unbond", body{Interface: interfaceID}, nil)
}

// AddBondMember adds an interface to a bond.
func (c *Client) AddBondMember(ctx context.Context, bondInterfaceID, memberInterfaceID string) error {
	type body struct {
		Interface string `json:"interface"`
		Bond      string `json:"bond"`
	}
	return c.call(ctx, "addbond", body{Interface: bondInterfaceID, Bond: memberInterfaceID}, nil)
}

// DeleteBondMember removes an interface from a bond.
func (c *Client) DeleteBondMember(ctx context.Context, bondInterfaceID, memberInterfaceID string) error {
	type body struct {
		Interface string `json:"interface"`
		Bond      string `json:"bond"`
	}
	return c.call(ctx, "delbond", body{Interface: bondInterfaceID, Bond: memberInterfaceID}, nil)
}

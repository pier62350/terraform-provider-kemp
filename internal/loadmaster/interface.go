// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type Interface struct {
	Id                  int32    `json:"Id"`
	IPAddress           string   `json:"IPAddress"`
	Mtu                 string   `json:"Mtu"`
	AdditionalAddresses []string `json:"AdditionalAddresses,omitempty"`
	InterfaceType       string   `json:"InterfaceType"`
	GeoTrafficEnable    bool     `json:"GeoTrafficEnable"`
	DefaultInterface    bool     `json:"DefaultInterface"`
}

type interfaceListResponse struct {
	Response
	Interface []Interface `json:"Interface,omitempty"`
}

func (c *Client) ShowInterface(ctx context.Context, id string) (*Interface, error) {
	type body struct {
		Interface string `json:"interface"`
	}
	var resp interfaceListResponse
	if err := c.call(ctx, "showiface", body{Interface: id}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Interface) == 0 {
		return nil, &Error{Message: "interface not found"}
	}
	return &resp.Interface[0], nil
}

func (c *Client) ListInterfaces(ctx context.Context) ([]Interface, error) {
	var resp interfaceListResponse
	if err := c.call(ctx, "showiface", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Interface, nil
}

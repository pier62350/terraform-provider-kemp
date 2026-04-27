// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// RealServer mirrors the addrs/showrs response shape.
type RealServer struct {
	RsIndex  int32  `json:"RsIndex"`
	VSIndex  int32  `json:"VSIndex"`
	Address  string `json:"Addr"`
	Port     int32  `json:"Port"`
	Weight   int32  `json:"Weight"`
	Forward  string `json:"Forward,omitempty"`
	Enable   *bool  `json:"Enable,omitempty"`
	Limit    int32  `json:"Limit,omitempty"`
	Critical *bool  `json:"Critical,omitempty"`
	Follow   int32  `json:"Follow,omitempty"`
	DnsName  string `json:"DnsName,omitempty"`
}

// RealServerParams are the modifiable knobs.
type RealServerParams struct {
	Weight   int32  `json:"Weight,omitempty"`
	Forward  string `json:"Forward,omitempty"`
	Enable   *bool  `json:"Enable,omitempty"`
	Limit    int32  `json:"Limit,omitempty"`
	Critical *bool  `json:"Critical,omitempty"`
	Follow   int32  `json:"Follow,omitempty"`
}

// rsListResponse is the wire shape: addrs/showrs return a list of RS rows
// even when the request was for a single one.
type rsListResponse struct {
	Response
	Rs []RealServer `json:"Rs"`
}

// AddRealServer attaches a backend to the given VS.
func (c *Client) AddRealServer(ctx context.Context, vsID, address, port string, p RealServerParams) (*RealServer, error) {
	type body struct {
		VS   string `json:"vs"`
		Rs   string `json:"rs"`
		RsP  string `json:"rsport"`
		RealServerParams
	}
	var resp rsListResponse
	if err := c.call(ctx, "addrs", body{VS: "!" + vsID, Rs: address, RsP: port, RealServerParams: p}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Rs) == 0 {
		return nil, &Error{Message: "addrs returned no rows"}
	}
	return &resp.Rs[len(resp.Rs)-1], nil
}

// ShowRealServer reads an RS by VS Index + RS Index.
func (c *Client) ShowRealServer(ctx context.Context, vsID, rsID string) (*RealServer, error) {
	type body struct {
		VS string `json:"vs"`
		Rs string `json:"rs"`
	}
	var resp rsListResponse
	if err := c.call(ctx, "showrs", body{VS: "!" + vsID, Rs: "!" + rsID}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Rs) == 0 {
		return nil, &Error{Message: "Unknown RS"}
	}
	return &resp.Rs[len(resp.Rs)-1], nil
}

// ModifyRealServer updates an RS in place.
func (c *Client) ModifyRealServer(ctx context.Context, vsID, rsID string, p RealServerParams) (*RealServer, error) {
	type body struct {
		VS string `json:"vs"`
		Rs string `json:"rs"`
		RealServerParams
	}
	var resp rsListResponse
	if err := c.call(ctx, "modrs", body{VS: "!" + vsID, Rs: "!" + rsID, RealServerParams: p}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Rs) == 0 {
		return nil, &Error{Message: "modrs returned no rows"}
	}
	return &resp.Rs[len(resp.Rs)-1], nil
}

// DeleteRealServer removes an RS.
func (c *Client) DeleteRealServer(ctx context.Context, vsID, rsID string) error {
	type body struct {
		VS string `json:"vs"`
		Rs string `json:"rs"`
	}
	return c.call(ctx, "delrs", body{VS: "!" + vsID, Rs: "!" + rsID}, nil)
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type LDAPEndpoint struct {
	Name      string `json:"name"`
	LDAPType  string `json:"ldaptype"`
	VInterval int32  `json:"vinterval"`
	Timeout   int32  `json:"timeout"`
}

type LDAPEndpointParams struct {
	Name          string `json:"name,omitempty"`
	LDAPType      string `json:"ldaptype,omitempty"`
	Server        string `json:"server,omitempty"`
	VInterval     *int32 `json:"vinterval,omitempty"`
	ReferralCount *int32 `json:"referralcount,omitempty"`
}

type ldapEndpointResponse struct {
	Response
	Data *LDAPEndpoint `json:"Data,omitempty"`
}

func (c *Client) AddLDAPEndpoint(ctx context.Context, name string, p LDAPEndpointParams) error {
	p.Name = name
	return c.call(ctx, "addldapendpoint", p, nil)
}

func (c *Client) ModifyLDAPEndpoint(ctx context.Context, name string, p LDAPEndpointParams) error {
	p.Name = name
	return c.call(ctx, "modifyldapendpoint", p, nil)
}

func (c *Client) ShowLDAPEndpoint(ctx context.Context, name string) (*LDAPEndpoint, error) {
	type body struct {
		Name string `json:"name"`
	}
	var resp ldapEndpointResponse
	if err := c.call(ctx, "showldapendpoint", body{Name: name}, &resp); err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, &Error{Message: "LDAP endpoint not found"}
	}
	return resp.Data, nil
}

func (c *Client) DeleteLDAPEndpoint(ctx context.Context, name string) error {
	type body struct {
		Name string `json:"name"`
	}
	return c.call(ctx, "deleteldapendpoint", body{Name: name}, nil)
}

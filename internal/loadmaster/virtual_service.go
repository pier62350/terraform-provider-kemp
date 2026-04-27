// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// VirtualService mirrors the relevant subset of fields returned by the
// addvs/showvs/modvs commands. Field names match the LoadMaster JSON
// (PascalCase) via explicit json tags.
type VirtualService struct {
	Index           int32  `json:"Index"`
	Address         string `json:"VSAddress"`
	Port            string `json:"VSPort"`
	Protocol        string `json:"Protocol"`
	VSType          string `json:"VStype"`
	NickName        string `json:"NickName"`
	Enable          *bool  `json:"Enable,omitempty"`
	SSLAcceleration *bool  `json:"SSLAcceleration,omitempty"`
	CertFile        string `json:"CertFile,omitempty"`

	// ESP (Edge Security Pack)
	EspEnabled          *bool  `json:"EspEnabled,omitempty"`
	AllowedHosts        string `json:"AllowedHosts,omitempty"`
	AllowedDirectories  string `json:"AllowedDirectories,omitempty"`
	InputAuthMode       string `json:"InputAuthMode,omitempty"`
	OutputAuthMode      string `json:"OutputAuthMode,omitempty"`
	IncludeNestedGroups *bool  `json:"IncludeNestedGroups,omitempty"`
	DisplayPubPriv      *bool  `json:"DisplayPubPriv,omitempty"`
	EspLogs             *bool  `json:"EspLogs,omitempty"`
}

// VirtualServiceParams are the optional knobs for create/modify.
// Only fields that are non-nil / non-empty get sent to LoadMaster.
type VirtualServiceParams struct {
	NickName        string `json:"NickName,omitempty"`
	VSType          string `json:"VStype,omitempty"`
	Enable          *bool  `json:"Enable,omitempty"`
	SSLAcceleration *bool  `json:"SSLAcceleration,omitempty"`
	CertFile        string `json:"CertFile,omitempty"`

	// ESP (Edge Security Pack)
	EspEnabled          *bool  `json:"EspEnabled,omitempty"`
	AllowedHosts        string `json:"AllowedHosts,omitempty"`
	AllowedDirectories  string `json:"AllowedDirectories,omitempty"`
	InputAuthMode       string `json:"InputAuthMode,omitempty"`
	OutputAuthMode      string `json:"OutputAuthMode,omitempty"`
	IncludeNestedGroups *bool  `json:"IncludeNestedGroups,omitempty"`
	DisplayPubPriv      *bool  `json:"DisplayPubPriv,omitempty"`
	EspLogs             *bool  `json:"EspLogs,omitempty"`
}

type vsResponse struct {
	Response
	VirtualService
}

// AddVirtualService creates a new virtual service.
func (c *Client) AddVirtualService(ctx context.Context, address, port, protocol string, p VirtualServiceParams) (*VirtualService, error) {
	type body struct {
		VS       string `json:"vs"`
		Port     string `json:"port"`
		Protocol string `json:"prot"`
		VirtualServiceParams
	}
	var resp vsResponse
	if err := c.call(ctx, "addvs", body{VS: address, Port: port, Protocol: protocol, VirtualServiceParams: p}, &resp); err != nil {
		return nil, err
	}
	return &resp.VirtualService, nil
}

// ShowVirtualService reads a single VS by its numeric Index.
//
// LoadMaster's parser interprets a numeric `vs` as the Index automatically;
// the "!N" prefix syntax is rejected by some firmware revs (it falls back to
// address-mode parsing and errors on missing `port`). Bare numeric Index is
// the safe form across versions.
func (c *Client) ShowVirtualService(ctx context.Context, id string) (*VirtualService, error) {
	type body struct {
		VS string `json:"vs"`
	}
	var resp vsResponse
	if err := c.call(ctx, "showvs", body{VS: id}, &resp); err != nil {
		return nil, err
	}
	return &resp.VirtualService, nil
}

// ModifyVirtualService updates a VS in place.
func (c *Client) ModifyVirtualService(ctx context.Context, id string, p VirtualServiceParams) (*VirtualService, error) {
	type body struct {
		VS string `json:"vs"`
		VirtualServiceParams
	}
	var resp vsResponse
	if err := c.call(ctx, "modvs", body{VS: id, VirtualServiceParams: p}, &resp); err != nil {
		return nil, err
	}
	return &resp.VirtualService, nil
}

// DeleteVirtualService removes a VS by Index.
func (c *Client) DeleteVirtualService(ctx context.Context, id string) error {
	type body struct {
		VS string `json:"vs"`
	}
	return c.call(ctx, "delvs", body{VS: id}, nil)
}

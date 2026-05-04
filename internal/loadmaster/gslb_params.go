// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"strings"
)

// GSLBParams holds miscellaneous GSLB/GEO DNS zone and SOA parameters.
type GSLBParams struct {
	Zone              string `json:"Zone"`
	SourceOfAuthority string `json:"SourceOfAuthority"`
	Namesrv           string `json:"Namesrv"`
	SOAEmail          string `json:"SOAEmail"`
	TTL               int32  `json:"TTL"`
	Persist           int32  `json:"Persist"`
}

type gslbParamsResponse struct {
	Response
	GSLBParams
}

// GetGSLBParams fetches the current GSLB miscellaneous parameters.
func (c *Client) GetGSLBParams(ctx context.Context) (*GSLBParams, error) {
	var resp gslbParamsResponse
	if err := c.call(ctx, "listparams", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.GSLBParams, nil
}

// SetGSLBParams updates the GSLB miscellaneous parameters.
func (c *Client) SetGSLBParams(ctx context.Context, p GSLBParams) error {
	type body struct {
		Zone              string `json:"zone,omitempty"`
		SourceOfAuthority string `json:"SourceOfAuthority,omitempty"`
		Namesrv           string `json:"namesrv,omitempty"`
		SOAEmail          string `json:"SOAEmail,omitempty"`
		TTL               int32  `json:"TTL,omitempty"`
		Persist           int32  `json:"persist,omitempty"`
	}
	return c.call(ctx, "modparams", body{
		Zone:              p.Zone,
		SourceOfAuthority: p.SourceOfAuthority,
		Namesrv:           p.Namesrv,
		SOAEmail:          p.SOAEmail,
		TTL:               p.TTL,
		Persist:           p.Persist,
	}, nil)
}

// GSLBIPRange represents a client IP address or CIDR registered for GEO routing.
type GSLBIPRange struct {
	IP   string `json:"IP"`
	Lat  int32  `json:"Lat,omitempty"`
	Long int32  `json:"Long,omitempty"`
}

type gslbIPRangeShowResponse struct {
	Response
	GSLBIPRange
}

type gslbIPRangeListResponse struct {
	Response
	IP []GSLBIPRange `json:"IP,omitempty"`
}

// AddGSLBIPRange registers a new IP/CIDR for GEO routing.
func (c *Client) AddGSLBIPRange(ctx context.Context, ip string) error {
	type body struct {
		IP string `json:"ip"`
	}
	return c.call(ctx, "addip", body{IP: ip}, nil)
}

// DeleteGSLBIPRange removes an IP/CIDR from GEO routing.
func (c *Client) DeleteGSLBIPRange(ctx context.Context, ip string) error {
	type body struct {
		IP string `json:"ip"`
	}
	return c.call(ctx, "delip", body{IP: ip}, nil)
}

// ShowGSLBIPRange returns the GEO routing details for a specific IP/CIDR.
func (c *Client) ShowGSLBIPRange(ctx context.Context, ip string) (*GSLBIPRange, error) {
	type body struct {
		IP string `json:"ip"`
	}
	var resp gslbIPRangeShowResponse
	if err := c.call(ctx, "showip", body{IP: ip}, &resp); err != nil {
		return nil, err
	}
	return &resp.GSLBIPRange, nil
}

// SetGSLBIPLocation sets the geographic coordinates for an IP/CIDR entry.
func (c *Client) SetGSLBIPLocation(ctx context.Context, ip string, lat, long int32) error {
	type body struct {
		IP   string `json:"ip"`
		Lat  int32  `json:"lat"`
		Long int32  `json:"long"`
	}
	return c.call(ctx, "modiploc", body{IP: ip, Lat: lat, Long: long}, nil)
}

// DeleteGSLBIPLocation removes the geographic coordinates from an IP/CIDR entry.
func (c *Client) DeleteGSLBIPLocation(ctx context.Context, ip string) error {
	type body struct {
		IP string `json:"ip"`
	}
	return c.call(ctx, "deliploc", body{IP: ip}, nil)
}

// gslbACLListResponse holds the response from geoacl.listcustom.
type gslbACLListResponse struct {
	Response
	Addr       []string `json:"Addr,omitempty"`
	CustomList []string `json:"CustomList,omitempty"`
}

// AddGSLBACLEntry adds an IP or CIDR to the GEO feature custom allow-list.
func (c *Client) AddGSLBACLEntry(ctx context.Context, addr string) error {
	type body struct {
		Addr string `json:"addr"`
	}
	return c.call(ctx, "geoacl.addcustom", body{Addr: addr}, nil)
}

// DeleteGSLBACLEntry removes an IP or CIDR from the GEO feature custom allow-list.
func (c *Client) DeleteGSLBACLEntry(ctx context.Context, addr string) error {
	type body struct {
		Addr string `json:"addr"`
	}
	return c.call(ctx, "geoacl.removecustom", body{Addr: addr}, nil)
}

// ListGSLBACLEntries returns all addresses in the GEO feature custom allow-list.
func (c *Client) ListGSLBACLEntries(ctx context.Context) ([]string, error) {
	var resp gslbACLListResponse
	if err := c.call(ctx, "geoacl.listcustom", nil, &resp); err != nil {
		return nil, err
	}
	// The API may return entries under "Addr" or "CustomList" — handle both.
	if len(resp.Addr) > 0 {
		return resp.Addr, nil
	}
	return resp.CustomList, nil
}

// IsGSLBIPNotFound reports whether an error from ShowGSLBIPRange indicates
// the entry does not exist on the LoadMaster.
func IsGSLBIPNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Unknown IP") ||
		strings.Contains(msg, "not found") ||
		strings.Contains(msg, "No such")
}

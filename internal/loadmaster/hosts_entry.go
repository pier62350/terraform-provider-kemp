// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type HostsEntry struct {
	HostIPAddress string `json:"HostIPAddress"`
	HostFqdn      string `json:"HostFqdn"`
}

type hostsEntryListResponse struct {
	Response
	HostsEntry []HostsEntry `json:"HostsEntry,omitempty"`
}

func (c *Client) AddHostsEntry(ctx context.Context, ip, fqdn string) error {
	type body struct {
		HostIP   string `json:"hostip"`
		HostFqdn string `json:"hostfqdn"`
	}
	return c.call(ctx, "addhostsentry", body{HostIP: ip, HostFqdn: fqdn}, nil)
}

func (c *Client) DeleteHostsEntry(ctx context.Context, ip string) error {
	type body struct {
		HostIP string `json:"hostip"`
	}
	return c.call(ctx, "delhostsentry", body{HostIP: ip}, nil)
}

func (c *Client) ListHostsEntries(ctx context.Context) ([]HostsEntry, error) {
	var resp hostsEntryListResponse
	if err := c.call(ctx, "gethosts", nil, &resp); err != nil {
		return nil, err
	}
	return resp.HostsEntry, nil
}

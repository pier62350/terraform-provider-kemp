// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"strings"
)

type RemoteGroup struct {
	Name  string `json:"Name"`
	Perms string `json:"Perms"`
}

type groupShowResponse struct {
	Response
	RemoteUserGroup RemoteGroup `json:"RemoteUserGroup"`
}

func (c *Client) AddRemoteGroup(ctx context.Context, group string) error {
	type body struct {
		Group string `json:"group"`
	}
	return c.call(ctx, "groupaddremote", body{Group: group}, nil)
}

func (c *Client) ShowRemoteGroup(ctx context.Context, group string) (*RemoteGroup, error) {
	type body struct {
		Group string `json:"group"`
	}
	var resp groupShowResponse
	if err := c.call(ctx, "groupshow", body{Group: group}, &resp); err != nil {
		return nil, err
	}
	resp.RemoteUserGroup.Perms = strings.TrimSpace(resp.RemoteUserGroup.Perms)
	return &resp.RemoteUserGroup, nil
}

func (c *Client) SetGroupPerms(ctx context.Context, group, perms string) error {
	type body struct {
		Group string `json:"group"`
		Perms string `json:"perms"`
	}
	return c.call(ctx, "groupsetperms", body{Group: group, Perms: perms}, nil)
}

func (c *Client) DeleteRemoteGroup(ctx context.Context, group string) error {
	type body struct {
		Group string `json:"group"`
	}
	return c.call(ctx, "groupdelremote", body{Group: group}, nil)
}

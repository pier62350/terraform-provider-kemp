// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type LocalUser struct {
	Name  string `json:"Name"`
	Perms string `json:"Perms"`
}

type localUserResponse struct {
	Response
	LocalUser
}

func (c *Client) AddLocalUser(ctx context.Context, username, password string) error {
	type body struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
	return c.call(ctx, "useraddlocal", body{User: username, Password: password}, nil)
}

func (c *Client) DeleteLocalUser(ctx context.Context, username string) error {
	type body struct {
		User string `json:"user"`
	}
	return c.call(ctx, "userdellocal", body{User: username}, nil)
}

func (c *Client) ShowLocalUser(ctx context.Context, username string) (*LocalUser, error) {
	type body struct {
		User string `json:"user"`
	}
	var resp localUserResponse
	if err := c.call(ctx, "usershow", body{User: username}, &resp); err != nil {
		return nil, err
	}
	return &resp.LocalUser, nil
}

func (c *Client) SetUserPerms(ctx context.Context, username, perms string) error {
	type body struct {
		User  string `json:"user"`
		Perms string `json:"perms"`
	}
	return c.call(ctx, "usersetperms", body{User: username, Perms: perms}, nil)
}

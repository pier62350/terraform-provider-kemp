// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type userCertDownloadResponse struct {
	Response
	Data string `json:"data"`
}

func (c *Client) NewUserCert(ctx context.Context, username, passphrase string) error {
	type body struct {
		User       string `json:"user"`
		Passphrase string `json:"passphrase,omitempty"`
	}
	return c.call(ctx, "usernewcert", body{User: username, Passphrase: passphrase}, nil)
}

// DownloadUserCert returns the PEM-encoded certificate and private key for a user.
func (c *Client) DownloadUserCert(ctx context.Context, username string) (string, error) {
	type body struct {
		User string `json:"user"`
	}
	var resp userCertDownloadResponse
	if err := c.call(ctx, "userdownloadcert", body{User: username}, &resp); err != nil {
		return "", err
	}
	return resp.Data, nil
}

func (c *Client) DeleteUserCert(ctx context.Context, username string) error {
	type body struct {
		User string `json:"user"`
	}
	return c.call(ctx, "userdelcert", body{User: username}, nil)
}

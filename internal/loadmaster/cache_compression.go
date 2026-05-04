// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// AddNoCacheExtension registers a file extension that must not be cached.
func (c *Client) AddNoCacheExtension(ctx context.Context, ext string) error {
	type body struct {
		Param string `json:"param"`
	}
	return c.call(ctx, "addnocache", body{Param: ext}, nil)
}

// DeleteNoCacheExtension removes a file extension from the no-cache list.
func (c *Client) DeleteNoCacheExtension(ctx context.Context, ext string) error {
	type body struct {
		Param string `json:"param"`
	}
	return c.call(ctx, "delnocache", body{Param: ext}, nil)
}

// AddNoCompressExtension registers a file extension that must not be compressed.
func (c *Client) AddNoCompressExtension(ctx context.Context, ext string) error {
	type body struct {
		Param string `json:"param"`
	}
	return c.call(ctx, "addnocompress", body{Param: ext}, nil)
}

// DeleteNoCompressExtension removes a file extension from the no-compress list.
func (c *Client) DeleteNoCompressExtension(ctx context.Context, ext string) error {
	type body struct {
		Param string `json:"param"`
	}
	return c.call(ctx, "delnocompress", body{Param: ext}, nil)
}

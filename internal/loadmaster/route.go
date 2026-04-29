// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type Route struct {
	Destination string `json:"Destination"`
	Gateway     string `json:"Gateway"`
}

type routeListResponse struct {
	Response
	Route []Route `json:"Route,omitempty"`
}

func (c *Client) AddRoute(ctx context.Context, dest, gateway string) error {
	type body struct {
		Dest    string `json:"dest"`
		Gateway string `json:"gateway"`
	}
	return c.call(ctx, "addroute", body{Dest: dest, Gateway: gateway}, nil)
}

func (c *Client) DeleteRoute(ctx context.Context, dest string) error {
	type body struct {
		Dest string `json:"dest"`
	}
	return c.call(ctx, "delroute", body{Dest: dest}, nil)
}

func (c *Client) ListRoutes(ctx context.Context) ([]Route, error) {
	var resp routeListResponse
	if err := c.call(ctx, "showroute", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Route, nil
}

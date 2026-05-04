// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type k8sValueResponse struct {
	Response
	Value string `json:"Value"`
}

// SetK8sConfig uploads a base64-encoded kubeconfig to the LoadMaster.
func (c *Client) SetK8sConfig(ctx context.Context, base64Data string) error {
	type body struct {
		Data string `json:"data"`
	}
	return c.call(ctx, "addlmingressk8sconf", body{Data: base64Data}, nil)
}

// DeleteK8sConfig removes the kubeconfig from the LoadMaster.
func (c *Client) DeleteK8sConfig(ctx context.Context) error {
	return c.call(ctx, "dellmingressk8sconf", nil, nil)
}

// GetK8sMode returns the current ingress controller mode.
func (c *Client) GetK8sMode(ctx context.Context) (string, error) {
	var resp k8sValueResponse
	if err := c.call(ctx, "getlmingressmode", nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// SetK8sMode sets the ingress controller mode (e.g. "active", "passive").
func (c *Client) SetK8sMode(ctx context.Context, mode string) error {
	type body struct {
		Mode string `json:"mode"`
	}
	return c.call(ctx, "setlmingressmode", body{Mode: mode}, nil)
}

// GetK8sNamespace returns the namespace the ingress controller watches.
func (c *Client) GetK8sNamespace(ctx context.Context) (string, error) {
	var resp k8sValueResponse
	if err := c.call(ctx, "getlmingressnamespace", nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// SetK8sNamespace sets the namespace the ingress controller watches.
func (c *Client) SetK8sNamespace(ctx context.Context, namespace string) error {
	type body struct {
		Namespace string `json:"namespace"`
	}
	return c.call(ctx, "setlmingressnamespace", body{Namespace: namespace}, nil)
}

// GetK8sWatchTimeout returns the current watch timeout value as a string.
func (c *Client) GetK8sWatchTimeout(ctx context.Context) (string, error) {
	var resp k8sValueResponse
	if err := c.call(ctx, "getlmingresswatchtimeout", nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// SetK8sWatchTimeout sets the watch timeout (in seconds, as a string).
func (c *Client) SetK8sWatchTimeout(ctx context.Context, timeout string) error {
	type body struct {
		WatchTimeout string `json:"watchtimeout"`
	}
	return c.call(ctx, "setlmingresswatchtimeout", body{WatchTimeout: timeout}, nil)
}

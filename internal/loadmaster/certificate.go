// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// CertInfo is the metadata Kemp returns for a stored certificate.
type CertInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Modulus string `json:"modulus,omitempty"`
}

type listCertResponse struct {
	Response
	Cert []CertInfo `json:"cert"`
}

// AddCertificate uploads a certificate. data must already be base64-encoded
// (PFX bundle or PEM text — Kemp accepts both, password is required only for
// encrypted PFX bundles).
func (c *Client) AddCertificate(ctx context.Context, name, base64Data, password string) error {
	type body struct {
		Cert     string  `json:"cert"`
		Data     string  `json:"data"`
		Password *string `json:"password,omitempty"`
	}
	var passPtr *string
	if password != "" {
		passPtr = &password
	}
	return c.call(ctx, "addcert", body{Cert: name, Data: base64Data, Password: passPtr}, nil)
}

// ListCertificates returns metadata for every cert known to LoadMaster.
func (c *Client) ListCertificates(ctx context.Context) ([]CertInfo, error) {
	var resp listCertResponse
	if err := c.call(ctx, "listcert", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Cert, nil
}

// FindCertificate returns the CertInfo for name, or nil if absent.
func (c *Client) FindCertificate(ctx context.Context, name string) (*CertInfo, error) {
	certs, err := c.ListCertificates(ctx)
	if err != nil {
		return nil, err
	}
	for i := range certs {
		if certs[i].Name == name {
			return &certs[i], nil
		}
	}
	return nil, nil
}

// DeleteCertificate removes a certificate by name.
func (c *Client) DeleteCertificate(ctx context.Context, name string) error {
	type body struct {
		Cert string `json:"cert"`
	}
	return c.call(ctx, "delcert", body{Cert: name}, nil)
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// ACMECertificateInfo mirrors the fields returned by getacmecert / listacmecert.
type ACMECertificateInfo struct {
	Identifier            string `json:"Identifier"`
	DomainName            string `json:"DomainName"`
	ExpiryDate            string `json:"ExpiryDate"`
	SubjectAlternateNames string `json:"SubjectAlternateNames,omitempty"`
	Type                  string `json:"Type"`
	KeySize               string `json:"KeySize"`
	HTTPChallengeVS       string `json:"HTTPChallengeVS,omitempty"`
	VirtualServices       string `json:"VirtualServices,omitempty"`
}

// AddACMECertParams carries the optional knobs for AddACMECertificate.
type AddACMECertParams struct {
	CommonName       string // required (cn)
	VirtualServiceID string // required (vid)
	ACMEType         string // required ("1"=Let's Encrypt, "2"=DigiCert)
	KeySize          int    // optional (default 2048)
	DNSAPI           string // optional — DNS-01 provider for wildcard certs (e.g. "godaddy.com")
	DNSAPIParams     string // optional — credentials for the DNS provider
	Email            string // optional — registration email
}

// AddACMECertificate requests a new ACME certificate. The cert is associated
// with a Virtual Service so LoadMaster can serve the HTTP-01 challenge there
// (or DNS-01 if dnsapi/dnsapiparams are set, for wildcards).
//
// Note: ACME issuance is asynchronous. The command returns once the request
// has been accepted; the cert may take seconds to minutes to actually be
// issued by the CA. Subsequent reads via GetACMECertificate will reflect
// progress.
func (c *Client) AddACMECertificate(ctx context.Context, name string, p AddACMECertParams) error {
	type body struct {
		Cert         string `json:"cert"`
		CN           string `json:"cn"`
		VID          string `json:"vid"`
		ACMEType     string `json:"acmetype"`
		KeySize      int    `json:"keysize,omitempty"`
		DNSAPI       string `json:"dnsapi,omitempty"`
		DNSAPIParams string `json:"dnsapiparams,omitempty"`
		Email        string `json:"email,omitempty"`
	}
	return c.call(ctx, "addacmecert", body{
		Cert:         name,
		CN:           p.CommonName,
		VID:          p.VirtualServiceID,
		ACMEType:     p.ACMEType,
		KeySize:      p.KeySize,
		DNSAPI:       p.DNSAPI,
		DNSAPIParams: p.DNSAPIParams,
		Email:        p.Email,
	}, nil)
}

type acmeCertResponse struct {
	Response
	Data ACMECertificateInfo `json:"Data"`
}

// GetACMECertificate reads one ACME cert's details. Returns nil if the cert
// is not present (use loadmaster.IsACMENotFound to test).
func (c *Client) GetACMECertificate(ctx context.Context, name, acmeType string) (*ACMECertificateInfo, error) {
	type body struct {
		Cert     string `json:"cert"`
		ACMEType string `json:"acmetype"`
	}
	var resp acmeCertResponse
	if err := c.call(ctx, "getacmecert", body{Cert: name, ACMEType: acmeType}, &resp); err != nil {
		return nil, err
	}
	if resp.Data.Identifier == "" {
		return nil, &Error{Message: "Unknown ACME cert"}
	}
	return &resp.Data, nil
}

// DeleteACMECertificate removes an ACME cert.
func (c *Client) DeleteACMECertificate(ctx context.Context, name, acmeType string) error {
	type body struct {
		Cert     string `json:"cert"`
		ACMEType string `json:"acmetype"`
	}
	return c.call(ctx, "delacmecert", body{Cert: name, ACMEType: acmeType}, nil)
}

// RenewACMECertificate triggers a renewal. Provider does not call this from
// CRUD lifecycle; exposed for future `kemp_acme_certificate_renewal` action
// or external orchestration.
func (c *Client) RenewACMECertificate(ctx context.Context, name, acmeType string) error {
	type body struct {
		Cert     string `json:"cert"`
		ACMEType string `json:"acmetype"`
	}
	return c.call(ctx, "renewacmecert", body{Cert: name, ACMEType: acmeType}, nil)
}

// ACMEAccountInfo holds the result of acmeaccountinfo.
type ACMEAccountInfo struct {
	AccountType      string `json:"AccountType"`
	AccountID        string `json:"AccountID"`
	AccountDirectory string `json:"AccountDirectory"`
	AccountEmail     string `json:"AccountEmail"`
}

type acmeAccountInfoResponse struct {
	Response
	Data ACMEAccountInfo `json:"Data"`
}

// GetACMEAccountInfo reads the registered ACME account for the given type.
func (c *Client) GetACMEAccountInfo(ctx context.Context, acmeType string) (*ACMEAccountInfo, error) {
	type body struct {
		ACMEType string `json:"acmetype"`
	}
	var resp acmeAccountInfoResponse
	if err := c.call(ctx, "acmeaccountinfo", body{ACMEType: acmeType}, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// RegisterACMEAccount registers a new account with the configured ACME
// provider. Email is optional (Let's Encrypt accepts blank); the provider
// will use whatever directory URL is currently set, so call
// SetACMEDirectoryURL first if you need to point at staging or DigiCert.
func (c *Client) RegisterACMEAccount(ctx context.Context, acmeType, email string) error {
	type body struct {
		ACMEType string `json:"acmetype"`
		Email    string `json:"email,omitempty"`
	}
	return c.call(ctx, "registeracmeaccount", body{ACMEType: acmeType, Email: email}, nil)
}

// SetACMEDirectoryURL configures which ACME endpoint LoadMaster talks to.
// For Let's Encrypt staging, that's
// https://acme-staging-v02.api.letsencrypt.org/directory.
func (c *Client) SetACMEDirectoryURL(ctx context.Context, acmeType, url string) error {
	type body struct {
		ACMEType     string `json:"acmetype"`
		DirectoryURL string `json:"directoryurl"`
	}
	return c.call(ctx, "setacmedirectoryurl", body{ACMEType: acmeType, DirectoryURL: url}, nil)
}

// GetACMEDirectoryURL returns the configured ACME endpoint.
func (c *Client) GetACMEDirectoryURL(ctx context.Context, acmeType string) (string, error) {
	type body struct {
		ACMEType string `json:"acmetype"`
	}
	type resp struct {
		Response
		DirectoryURL string `json:"directoryurl"`
	}
	var out resp
	if err := c.call(ctx, "getacmedirectoryurl", body{ACMEType: acmeType}, &out); err != nil {
		return "", err
	}
	return out.DirectoryURL, nil
}

// SetACMERenewPeriod sets the days-before-expiry threshold at which
// LoadMaster auto-renews ACME certs (1-60).
func (c *Client) SetACMERenewPeriod(ctx context.Context, acmeType string, days int32) error {
	type body struct {
		ACMEType    string `json:"acmetype"`
		RenewPeriod int32  `json:"renewperiod"`
	}
	return c.call(ctx, "setacmerenewperiod", body{ACMEType: acmeType, RenewPeriod: days}, nil)
}

// GetACMERenewPeriod returns the configured renew period.
func (c *Client) GetACMERenewPeriod(ctx context.Context, acmeType string) (int32, error) {
	type body struct {
		ACMEType string `json:"acmetype"`
	}
	type resp struct {
		Response
		RenewPeriod int32 `json:"renewperiod"`
	}
	var out resp
	if err := c.call(ctx, "getacmerenewperiod", body{ACMEType: acmeType}, &out); err != nil {
		return 0, err
	}
	return out.RenewPeriod, nil
}

// DeleteACMEConfig wipes the ACME configuration for the given provider type
// (only succeeds if no ACME certs exist for that type).
func (c *Client) DeleteACMEConfig(ctx context.Context, acmeType string) error {
	type body struct {
		ACMEType string `json:"acmetype"`
	}
	return c.call(ctx, "delacmeconfig", body{ACMEType: acmeType}, nil)
}

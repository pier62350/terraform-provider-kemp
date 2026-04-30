// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// SSODomain mirrors the fields returned by showdomain / moddomain.
// Wire format uses snake_case for most fields (unusual for the Kemp API).
type SSODomain struct {
	Id                 int32  `json:"Id"`
	Name               string `json:"Name"`
	AuthType           string `json:"auth_type,omitempty"`
	LDAPEndpoint       string `json:"ldap_endpoint,omitempty"`
	LDAPEpHC           int32  `json:"ldapephc,omitempty"`
	LDAPVersion        int32  `json:"ldap_version,omitempty"`
	ServerSide         string `json:"server_side,omitempty"`
	LogonFmt           string `json:"logon_fmt,omitempty"`
	LogonFmt2          string `json:"logon_fmt2,omitempty"`
	LogonDomain        string `json:"logon_domain,omitempty"`
	LogonTranscode     int32  `json:"logon_transcode,omitempty"`
	UserAccControl     int32  `json:"user_acc_control,omitempty"`
	MaxFailedAuths     int32  `json:"max_failed_auths,omitempty"`
	SessToutIdlePub    int32  `json:"sess_tout_idle_pub,omitempty"`
	SessToutDurPub     int32  `json:"sess_tout_duration_pub,omitempty"`
	SessToutIdlePriv   int32  `json:"sess_tout_idle_priv,omitempty"`
	SessToutDurPriv    int32  `json:"sess_tout_duration_priv,omitempty"`
	SessToutType       string `json:"sess_tout_type,omitempty"`
	ResetFailTout      int32  `json:"reset_fail_tout,omitempty"`
	UnblockTout        int32  `json:"unblock_tout,omitempty"`
	RadiusSendNasId    string `json:"radius_send_nas_id,omitempty"`
	RadiusNasId        string `json:"radius_nas_id,omitempty"`
	KerberosDomain     string `json:"kerberos_domain,omitempty"`
	KerberosKDC        string `json:"kerberos_kdc,omitempty"`
	KCDUsername        string `json:"kcd_username,omitempty"`
	LDAPAdmin          string `json:"ldap_admin,omitempty"`
	CertCheckASI       string `json:"cert_asi,omitempty"`
	CertCheckCN        string `json:"cert_check_cn,omitempty"`
	IdpEntityId        string `json:"idp_entity_id,omitempty"`
	IdpSsoUrl          string `json:"idp_sso_url,omitempty"`
	IdpLogoffUrl       string `json:"idp_logoff_url,omitempty"`
	IdpCert            string `json:"idp_cert,omitempty"`
	IdpMatchCert       string `json:"idp_match_cert,omitempty"`
	SpEntityId         string `json:"sp_entity_id,omitempty"`
	SpCert             string `json:"sp_cert,omitempty"`
	OidcAppId          string `json:"oidc_app_id,omitempty"`
	OidcRedirectUri    string `json:"oidc_redirect_uri,omitempty"`
	OidcAuthEpUrl      string `json:"oidc_auth_ep_url,omitempty"`
	OidcTokenEpUrl     string `json:"oidc_token_ep_url,omitempty"`
	OidcLogoffUrl      string `json:"oidc_logoff_url,omitempty"`
}

// SSODomainParams holds optional knobs passed to moddomain.
// Sensitive write-only fields (passwords) are included here but never appear in SSODomain reads.
type SSODomainParams struct {
	AuthType            string `json:"auth_type,omitempty"`
	LDAPEndpoint        string `json:"ldap_endpoint,omitempty"`
	LDAPEpHC            *int32 `json:"ldapephc,omitempty"`
	ServerSide          *bool  `json:"server_side,omitempty"`
	LogonFmt            string `json:"logon_fmt,omitempty"`
	LogonFmt2           string `json:"logon_fmt2,omitempty"`
	LogonDomain         string `json:"logon_domain,omitempty"`
	LogonTranscode      *bool  `json:"logon_transcode,omitempty"`
	UserAccControl      *int32 `json:"user_acc_control,omitempty"`
	MaxFailedAuths      *int32 `json:"max_failed_auths,omitempty"`
	SessToutIdlePub     *int32 `json:"sess_tout_idle_pub,omitempty"`
	SessToutDurPub      *int32 `json:"sess_tout_duration_pub,omitempty"`
	SessToutIdlePriv    *int32 `json:"sess_tout_idle_priv,omitempty"`
	SessToutDurPriv     *int32 `json:"sess_tout_duration_priv,omitempty"`
	SessToutType        string `json:"sess_tout_type,omitempty"`
	ResetFailTout       *int32 `json:"reset_fail_tout,omitempty"`
	UnblockTout         *int32 `json:"unblock_tout,omitempty"`
	Server              string `json:"server,omitempty"`
	RadiusSharedSecret  string `json:"radius_shared_secret,omitempty"`
	RadiusSendNasId     *bool  `json:"radius_send_nas_id,omitempty"`
	RadiusNasId         string `json:"radius_nas_id,omitempty"`
	KerberosDomain      string `json:"kerberos_domain,omitempty"`
	KerberosKDC         string `json:"kerberos_kdc,omitempty"`
	KCDUsername         string `json:"kcd_username,omitempty"`
	KCDPassword         string `json:"kcd_password,omitempty"`
	LDAPAdmin           string `json:"ldap_admin,omitempty"`
	LDAPPassword        string `json:"ldap_password,omitempty"`
	CertCheckASI        *bool  `json:"cert_check_asi,omitempty"`
	CertCheckCN         *bool  `json:"cert_check_cn,omitempty"`
	IdpEntityId         string `json:"idp_entity_id,omitempty"`
	IdpSsoUrl           string `json:"idp_sso_url,omitempty"`
	IdpLogoffUrl        string `json:"idp_logoff_url,omitempty"`
	IdpCert             string `json:"idp_cert,omitempty"`
	IdpMatchCert        *bool  `json:"idp_match_cert,omitempty"`
	SpEntityId          string `json:"sp_entity_id,omitempty"`
	SpCert              string `json:"sp_cert,omitempty"`
	OidcAppId           string `json:"oidc_app_id,omitempty"`
	OidcRedirectUri     string `json:"oidc_redirect_uri,omitempty"`
	OidcAuthEpUrl       string `json:"oidc_auth_ep_url,omitempty"`
	OidcTokenEpUrl      string `json:"oidc_token_ep_url,omitempty"`
	OidcLogoffUrl       string `json:"oidc_logoff_url,omitempty"`
	OidcSecret          string `json:"oidc_secret,omitempty"`
}

// showDomainResponse wraps the Domain array returned by showdomain / moddomain.
type showDomainResponse struct {
	Response
	Domain []SSODomain `json:"Domain"`
}

// addDomainResponse is the minimal response from adddomain (just code/status/message).
type addDomainResponse struct {
	Response
}

// AddSSODomain creates a new SSO domain by name, then applies params.
func (c *Client) AddSSODomain(ctx context.Context, name string, p SSODomainParams) (*SSODomain, error) {
	type addBody struct {
		Domain string `json:"domain"`
	}
	var ar addDomainResponse
	if err := c.call(ctx, "adddomain", addBody{Domain: name}, &ar); err != nil {
		return nil, err
	}
	return c.ModifySSODomain(ctx, name, p)
}

// ModifySSODomain updates an existing SSO domain.
func (c *Client) ModifySSODomain(ctx context.Context, name string, p SSODomainParams) (*SSODomain, error) {
	type modBody struct {
		Domain string `json:"domain"`
		SSODomainParams
	}
	var resp showDomainResponse
	if err := c.call(ctx, "moddomain", modBody{Domain: name, SSODomainParams: p}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Domain) == 0 {
		return nil, &Error{Code: 404, Message: "domain not found in response"}
	}
	return &resp.Domain[0], nil
}

// ShowSSODomain reads a single SSO domain by name.
func (c *Client) ShowSSODomain(ctx context.Context, name string) (*SSODomain, error) {
	type body struct {
		Domain string `json:"domain"`
	}
	var resp showDomainResponse
	if err := c.call(ctx, "showdomain", body{Domain: name}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Domain) == 0 {
		return nil, &Error{Code: 404, Message: "domain not found"}
	}
	return &resp.Domain[0], nil
}

// DeleteSSODomain removes an SSO domain by name.
func (c *Client) DeleteSSODomain(ctx context.Context, name string) error {
	type body struct {
		Domain string `json:"domain"`
	}
	return c.call(ctx, "deldomain", body{Domain: name}, nil)
}

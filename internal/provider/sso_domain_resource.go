// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &SSODomainResource{}
	_ resource.ResourceWithImportState = &SSODomainResource{}
)

func NewSSODomainResource() resource.Resource { return &SSODomainResource{} }

type SSODomainResource struct{ client *loadmaster.Client }

type SSODomainResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	AuthType types.String `tfsdk:"auth_type"`

	// LDAP
	LDAPEndpoint types.String `tfsdk:"ldap_endpoint"`
	LDAPEpHC     types.Bool   `tfsdk:"ldap_endpoint_health_check"`

	// RADIUS
	RadiusSharedSecret types.String `tfsdk:"radius_shared_secret"`
	RadiusSendNasId    types.Bool   `tfsdk:"radius_send_nas_id"`
	RadiusNasId        types.String `tfsdk:"radius_nas_id"`

	// Logon options (LDAP / RADIUS)
	LogonFmt       types.String `tfsdk:"logon_fmt"`
	LogonFmt2      types.String `tfsdk:"logon_fmt2"`
	LogonDomain    types.String `tfsdk:"logon_domain"`
	LogonTranscode types.Bool   `tfsdk:"logon_transcode"`
	Server         types.String `tfsdk:"server"`

	// Session / lockout
	UserAccControl  types.Int64  `tfsdk:"user_acc_control"`
	MaxFailedAuths  types.Int64  `tfsdk:"max_failed_auths"`
	ResetFailTout   types.Int64  `tfsdk:"reset_fail_tout"`
	UnblockTout     types.Int64  `tfsdk:"unblock_tout"`
	SessToutIdlePub types.Int64  `tfsdk:"sess_tout_idle_pub"`
	SessToutDurPub  types.Int64  `tfsdk:"sess_tout_duration_pub"`
	SessToutIdlePriv types.Int64 `tfsdk:"sess_tout_idle_priv"`
	SessToutDurPriv types.Int64  `tfsdk:"sess_tout_duration_priv"`
	SessToutType    types.String `tfsdk:"sess_tout_type"`

	// KCD (Kerberos Constrained Delegation)
	ServerSide     types.Bool   `tfsdk:"server_side"`
	KerberosDomain types.String `tfsdk:"kerberos_domain"`
	KerberosKDC    types.String `tfsdk:"kerberos_kdc"`
	KCDUsername    types.String `tfsdk:"kcd_username"`
	KCDPassword    types.String `tfsdk:"kcd_password"`

	// Certificate auth
	LDAPAdmin    types.String `tfsdk:"ldap_admin"`
	LDAPPassword types.String `tfsdk:"ldap_password"`
	CertCheckASI types.Bool   `tfsdk:"cert_check_asi"`
	CertCheckCN  types.Bool   `tfsdk:"cert_check_cn"`

	// SAML
	IdpEntityId  types.String `tfsdk:"idp_entity_id"`
	IdpSsoUrl    types.String `tfsdk:"idp_sso_url"`
	IdpLogoffUrl types.String `tfsdk:"idp_logoff_url"`
	IdpCert      types.String `tfsdk:"idp_cert"`
	IdpMatchCert types.Bool   `tfsdk:"idp_match_cert"`
	SpEntityId   types.String `tfsdk:"sp_entity_id"`
	SpCert       types.String `tfsdk:"sp_cert"`

	// OIDC / OAuth
	OidcAppId      types.String `tfsdk:"oidc_app_id"`
	OidcRedirectUri types.String `tfsdk:"oidc_redirect_uri"`
	OidcAuthEpUrl  types.String `tfsdk:"oidc_auth_ep_url"`
	OidcTokenEpUrl types.String `tfsdk:"oidc_token_ep_url"`
	OidcLogoffUrl  types.String `tfsdk:"oidc_logoff_url"`
	OidcSecret     types.String `tfsdk:"oidc_secret"`
}

func (r *SSODomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sso_domain"
}

func (r *SSODomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Kemp LoadMaster SSO (Edge Security Pack) authentication domain.

An SSO domain defines how the LoadMaster authenticates users before forwarding requests to real servers. Supported authentication protocols: LDAP (Unencrypted / StartTLS / LDAPS), RADIUS, RSA SecurID, KCD (Kerberos Constrained Delegation), Certificates, SAML, and OIDC/OAuth.

Once created, reference the domain name in ` + "`kemp_virtual_service`" + ` or ` + "`kemp_sub_virtual_service`" + ` via ` + "`esp_sso_domain`" + ` (client-side) or ` + "`esp_sso_out_domain`" + ` (server-side KCD).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Domain name (same as `name`). Used as the Terraform resource ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Domain name (e.g. `example.com`). Case-insensitive on the LoadMaster; stored uppercased. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"auth_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Authentication protocol. One of: `LDAP-Unencrypted`, `LDAP-StartTLS` (default), `LDAP-LDAPS`, `RADIUS`, `RSA-SECURID`, `KCD`, `Certificates`, `RADIUS and LDAP-Unencrypted`, `RADIUS and LDAP-StartTLS`, `RADIUS and LDAP-LDAPS`, `RSA-SECURID and LDAP-Unencrypted`, `RSA-SECURID and LDAP-StartTLS`, `RSA-SECURID and LDAP-LDAPS`, `OIDC-OAUTH`.",
			},

			// LDAP
			"ldap_endpoint": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Name of an existing `kemp_config_ldap_endpoint` to use for LDAP authentication.",
			},
			"ldap_endpoint_health_check": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Use the LDAP endpoint admin credentials for health checks. Default: `true`.",
			},

			// RADIUS
			"radius_shared_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Optional. Shared secret between the LoadMaster and the RADIUS server. Write-only — not returned on read.",
			},
			"radius_send_nas_id": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Send a NAS identifier in RADIUS access requests. Default: `false`.",
			},
			"radius_nas_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Custom NAS identifier string. Defaults to the LoadMaster hostname when `radius_send_nas_id = true`.",
			},

			// Logon options
			"logon_fmt": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Logon string format for LDAP/RADIUS: `Principalname` (default), `Username`, `Username only` (RADIUS/RSA-SecurID only).",
			},
			"logon_fmt2": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Secondary logon format: `Principalname`, `Username`.",
			},
			"logon_domain": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Domain/Realm appended to the username during authentication (e.g. `CORP` for `user\\CORP` or `user@CORP`).",
			},
			"logon_transcode": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Transcode logon credentials from ISO-8859-1 to UTF-8. Default: `false`.",
			},
			"server": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. IP address of the authentication server(s). Space-separated for multiple addresses. Used when not referencing an LDAP endpoint.",
			},

			// Session / lockout
			"user_acc_control": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Interval (minutes, `0`–`300`) for periodic User Access Control checks. `0` = disabled. Default: `0`.",
			},
			"max_failed_auths": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Maximum failed login attempts before the user is locked out. `0` = never lock out. Default: `0`.",
			},
			"reset_fail_tout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Seconds before the failed-login counter resets. Range: `60`–`86400`. Default: `60`.",
			},
			"unblock_tout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Seconds before a locked-out account is automatically unblocked. Must be greater than `reset_fail_tout`. Range: `60`–`86400`. Default: `1800`.",
			},
			"sess_tout_idle_pub": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Session idle timeout in seconds (public/untrusted environment). Range: `60`–`604800`. Default: `900`.",
			},
			"sess_tout_duration_pub": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Maximum session duration in seconds (public/untrusted environment). Range: `60`–`604800`. Default: `1800`.",
			},
			"sess_tout_idle_priv": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Session idle timeout in seconds (private/trusted environment). Range: `60`–`604800`. Default: `900`.",
			},
			"sess_tout_duration_priv": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Maximum session duration in seconds (private/trusted environment). Range: `60`–`604800`. Default: `28800`.",
			},
			"sess_tout_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Session timeout enforcement: `idle time` (default) or `max duration`.",
			},

			// KCD
			"server_side": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. `true` = outbound KCD domain (server-side); `false` = inbound (client-side). Default: `false`.",
			},
			"kerberos_domain": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Kerberos realm (e.g. `CORP.EXAMPLE.COM`). Required for `auth_type = \"KCD\"`.",
			},
			"kerberos_kdc": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Address of the Kerberos Key Distribution Center (KDC). Required for `auth_type = \"KCD\"`.",
			},
			"kcd_username": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Service account username for KCD. Do not use quotes in the value.",
			},
			"kcd_password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Optional. Password for the KCD service account. Write-only — not returned on read. Do not use quotes in the value.",
			},

			// Certificate auth
			"ldap_admin": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. LDAP admin username for certificate-to-LDAP cross-check (`auth_type = \"Certificates\"`).",
			},
			"ldap_password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Optional. Password for the LDAP admin account. Write-only — not returned on read.",
			},
			"cert_check_asi": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Check the client certificate against the `altSecurityIdentities` attribute in AD. Only available when `auth_type = \"Certificates\"`. Default: `false`.",
			},
			"cert_check_cn": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Fall back to the certificate Common Name (CN) when SAN is unavailable. Default: `false`.",
			},

			// SAML
			"idp_entity_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Identity Provider (IdP) Entity ID for SAML. Max 255 characters.",
			},
			"idp_sso_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. IdP Single Sign-On URL for SAML.",
			},
			"idp_logoff_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. IdP Logoff URL for SAML. Max 255 characters.",
			},
			"idp_cert": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Name of the certificate used to verify the IdP SAML response.",
			},
			"idp_match_cert": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Require the IdP certificate to match the one in the SAML response. Default: `false`.",
			},
			"sp_entity_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Service Provider (SP) Entity ID shared with the IdP. Max 255 characters.",
			},
			"sp_cert": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Certificate used to sign SAML logoff requests. Set to `useselfsigned` (default) to use the LoadMaster self-signed cert.",
			},

			// OIDC / OAuth
			"oidc_app_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Application (client) ID for OIDC/OAuth. Max 255 characters.",
			},
			"oidc_redirect_uri": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Redirect URI(s) (reply URLs) for OIDC/OAuth. Space-separated for multiple. Max 255 characters.",
			},
			"oidc_auth_ep_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Authorization endpoint URL for OIDC/OAuth.",
			},
			"oidc_token_ep_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Token endpoint URL for OIDC/OAuth.",
			},
			"oidc_logoff_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Logoff URL for OIDC/OAuth.",
			},
			"oidc_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Optional. Client secret for the OIDC application. Write-only — not returned on read.",
			},
		},
	}
}

func (r *SSODomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*loadmaster.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configure type", fmt.Sprintf("Expected *loadmaster.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *SSODomainResource) paramsFromModel(m SSODomainResourceModel) loadmaster.SSODomainParams {
	p := loadmaster.SSODomainParams{
		AuthType:        m.AuthType.ValueString(),
		LDAPEndpoint:    m.LDAPEndpoint.ValueString(),
		LogonFmt:        m.LogonFmt.ValueString(),
		LogonFmt2:       m.LogonFmt2.ValueString(),
		LogonDomain:     m.LogonDomain.ValueString(),
		Server:          m.Server.ValueString(),
		KerberosDomain:  m.KerberosDomain.ValueString(),
		KerberosKDC:     m.KerberosKDC.ValueString(),
		KCDUsername:     m.KCDUsername.ValueString(),
		KCDPassword:     m.KCDPassword.ValueString(),
		LDAPAdmin:       m.LDAPAdmin.ValueString(),
		LDAPPassword:    m.LDAPPassword.ValueString(),
		IdpEntityId:     m.IdpEntityId.ValueString(),
		IdpSsoUrl:       m.IdpSsoUrl.ValueString(),
		IdpLogoffUrl:    m.IdpLogoffUrl.ValueString(),
		IdpCert:         m.IdpCert.ValueString(),
		SpEntityId:      m.SpEntityId.ValueString(),
		SpCert:          m.SpCert.ValueString(),
		OidcAppId:       m.OidcAppId.ValueString(),
		OidcRedirectUri: m.OidcRedirectUri.ValueString(),
		OidcAuthEpUrl:   m.OidcAuthEpUrl.ValueString(),
		OidcTokenEpUrl:  m.OidcTokenEpUrl.ValueString(),
		OidcLogoffUrl:   m.OidcLogoffUrl.ValueString(),
		OidcSecret:      m.OidcSecret.ValueString(),
		RadiusSharedSecret: m.RadiusSharedSecret.ValueString(),
		RadiusNasId:     m.RadiusNasId.ValueString(),
		SessToutType:    m.SessToutType.ValueString(),
	}
	if !m.LDAPEpHC.IsNull() && !m.LDAPEpHC.IsUnknown() {
		v := int32(0)
		if m.LDAPEpHC.ValueBool() {
			v = 1
		}
		p.LDAPEpHC = &v
	}
	if !m.RadiusSendNasId.IsNull() && !m.RadiusSendNasId.IsUnknown() {
		p.RadiusSendNasId = boolPtr(m.RadiusSendNasId.ValueBool())
	}
	if !m.LogonTranscode.IsNull() && !m.LogonTranscode.IsUnknown() {
		p.LogonTranscode = boolPtr(m.LogonTranscode.ValueBool())
	}
	if !m.ServerSide.IsNull() && !m.ServerSide.IsUnknown() {
		p.ServerSide = boolPtr(m.ServerSide.ValueBool())
	}
	if !m.CertCheckASI.IsNull() && !m.CertCheckASI.IsUnknown() {
		p.CertCheckASI = boolPtr(m.CertCheckASI.ValueBool())
	}
	if !m.CertCheckCN.IsNull() && !m.CertCheckCN.IsUnknown() {
		p.CertCheckCN = boolPtr(m.CertCheckCN.ValueBool())
	}
	if !m.IdpMatchCert.IsNull() && !m.IdpMatchCert.IsUnknown() {
		p.IdpMatchCert = boolPtr(m.IdpMatchCert.ValueBool())
	}
	if !m.UserAccControl.IsNull() && !m.UserAccControl.IsUnknown() {
		v := int32(m.UserAccControl.ValueInt64())
		p.UserAccControl = &v
	}
	if !m.MaxFailedAuths.IsNull() && !m.MaxFailedAuths.IsUnknown() {
		v := int32(m.MaxFailedAuths.ValueInt64())
		p.MaxFailedAuths = &v
	}
	if !m.ResetFailTout.IsNull() && !m.ResetFailTout.IsUnknown() {
		v := int32(m.ResetFailTout.ValueInt64())
		p.ResetFailTout = &v
	}
	if !m.UnblockTout.IsNull() && !m.UnblockTout.IsUnknown() {
		v := int32(m.UnblockTout.ValueInt64())
		p.UnblockTout = &v
	}
	if !m.SessToutIdlePub.IsNull() && !m.SessToutIdlePub.IsUnknown() {
		v := int32(m.SessToutIdlePub.ValueInt64())
		p.SessToutIdlePub = &v
	}
	if !m.SessToutDurPub.IsNull() && !m.SessToutDurPub.IsUnknown() {
		v := int32(m.SessToutDurPub.ValueInt64())
		p.SessToutDurPub = &v
	}
	if !m.SessToutIdlePriv.IsNull() && !m.SessToutIdlePriv.IsUnknown() {
		v := int32(m.SessToutIdlePriv.ValueInt64())
		p.SessToutIdlePriv = &v
	}
	if !m.SessToutDurPriv.IsNull() && !m.SessToutDurPriv.IsUnknown() {
		v := int32(m.SessToutDurPriv.ValueInt64())
		p.SessToutDurPriv = &v
	}
	return p
}

func (r *SSODomainResource) writeState(d *loadmaster.SSODomain, m *SSODomainResourceModel) {
	m.Id = types.StringValue(strings.ToLower(d.Name))
	m.Name = types.StringValue(strings.ToLower(d.Name))
	m.AuthType = types.StringValue(d.AuthType)
	m.LDAPEndpoint = types.StringValue(d.LDAPEndpoint)
	m.LDAPEpHC = types.BoolValue(d.LDAPEpHC != 0)
	m.LogonFmt = types.StringValue(d.LogonFmt)
	m.LogonFmt2 = types.StringValue(d.LogonFmt2)
	m.LogonDomain = types.StringValue(d.LogonDomain)
	m.LogonTranscode = types.BoolValue(d.LogonTranscode != 0)
	m.RadiusSendNasId = types.BoolValue(d.RadiusSendNasId == "1")
	m.RadiusNasId = types.StringValue(d.RadiusNasId)
	m.UserAccControl = types.Int64Value(int64(d.UserAccControl))
	m.MaxFailedAuths = types.Int64Value(int64(d.MaxFailedAuths))
	m.ResetFailTout = types.Int64Value(int64(d.ResetFailTout))
	m.UnblockTout = types.Int64Value(int64(d.UnblockTout))
	m.SessToutIdlePub = types.Int64Value(int64(d.SessToutIdlePub))
	m.SessToutDurPub = types.Int64Value(int64(d.SessToutDurPub))
	m.SessToutIdlePriv = types.Int64Value(int64(d.SessToutIdlePriv))
	m.SessToutDurPriv = types.Int64Value(int64(d.SessToutDurPriv))
	m.SessToutType = types.StringValue(d.SessToutType)
	m.ServerSide = types.BoolValue(d.ServerSide == "1")
	m.KerberosDomain = types.StringValue(d.KerberosDomain)
	m.KerberosKDC = types.StringValue(d.KerberosKDC)
	m.KCDUsername = types.StringValue(d.KCDUsername)
	m.LDAPAdmin = types.StringValue(d.LDAPAdmin)
	m.CertCheckASI = types.BoolValue(d.CertCheckASI != "" && d.CertCheckASI != "Not Specified" && d.CertCheckASI != "0")
	m.CertCheckCN = types.BoolValue(d.CertCheckCN == "1")
	m.IdpEntityId = types.StringValue(d.IdpEntityId)
	m.IdpSsoUrl = types.StringValue(d.IdpSsoUrl)
	m.IdpLogoffUrl = types.StringValue(d.IdpLogoffUrl)
	m.IdpCert = types.StringValue(d.IdpCert)
	m.IdpMatchCert = types.BoolValue(d.IdpMatchCert == "1")
	m.SpEntityId = types.StringValue(d.SpEntityId)
	m.SpCert = types.StringValue(d.SpCert)
	m.OidcAppId = types.StringValue(d.OidcAppId)
	m.OidcRedirectUri = types.StringValue(d.OidcRedirectUri)
	m.OidcAuthEpUrl = types.StringValue(d.OidcAuthEpUrl)
	m.OidcTokenEpUrl = types.StringValue(d.OidcTokenEpUrl)
	m.OidcLogoffUrl = types.StringValue(d.OidcLogoffUrl)
	// Write-only fields: RadiusSharedSecret, KCDPassword, LDAPPassword, OidcSecret
	// are not returned by the API — preserve whatever is in state.
}

func (r *SSODomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SSODomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := r.paramsFromModel(data)
	d, err := r.client.AddSSODomain(ctx, data.Name.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SSO domain", err.Error())
		return
	}

	r.writeState(d, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSODomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SSODomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.client.ShowSSODomain(ctx, data.Id.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SSO domain", err.Error())
		return
	}

	r.writeState(d, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSODomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SSODomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := r.paramsFromModel(data)
	d, err := r.client.ModifySSODomain(ctx, data.Id.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SSO domain", err.Error())
		return
	}

	r.writeState(d, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSODomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SSODomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSSODomain(ctx, data.Id.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting SSO domain", err.Error())
	}
}

func (r *SSODomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	d, err := r.client.ShowSSODomain(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing SSO domain", err.Error())
		return
	}

	var data SSODomainResourceModel
	r.writeState(d, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

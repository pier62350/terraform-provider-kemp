// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var _ datasource.DataSource = &SSODomainDataSource{}

func NewSSODomainDataSource() datasource.DataSource { return &SSODomainDataSource{} }

type SSODomainDataSource struct{ client *loadmaster.Client }

type SSODomainDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	AuthType types.String `tfsdk:"auth_type"`

	LDAPEndpoint types.String `tfsdk:"ldap_endpoint"`
	LDAPEpHC     types.Bool   `tfsdk:"ldap_endpoint_health_check"`

	RadiusSendNasId types.Bool   `tfsdk:"radius_send_nas_id"`
	RadiusNasId     types.String `tfsdk:"radius_nas_id"`

	LogonFmt       types.String `tfsdk:"logon_fmt"`
	LogonFmt2      types.String `tfsdk:"logon_fmt2"`
	LogonDomain    types.String `tfsdk:"logon_domain"`
	LogonTranscode types.Bool   `tfsdk:"logon_transcode"`
	Server         types.String `tfsdk:"server"`

	UserAccControl  types.Int64  `tfsdk:"user_acc_control"`
	MaxFailedAuths  types.Int64  `tfsdk:"max_failed_auths"`
	ResetFailTout   types.Int64  `tfsdk:"reset_fail_tout"`
	UnblockTout     types.Int64  `tfsdk:"unblock_tout"`
	SessToutIdlePub  types.Int64  `tfsdk:"sess_tout_idle_pub"`
	SessToutDurPub   types.Int64  `tfsdk:"sess_tout_duration_pub"`
	SessToutIdlePriv types.Int64  `tfsdk:"sess_tout_idle_priv"`
	SessToutDurPriv  types.Int64  `tfsdk:"sess_tout_duration_priv"`
	SessToutType     types.String `tfsdk:"sess_tout_type"`

	ServerSide     types.Bool   `tfsdk:"server_side"`
	KerberosDomain types.String `tfsdk:"kerberos_domain"`
	KerberosKDC    types.String `tfsdk:"kerberos_kdc"`
	KCDUsername    types.String `tfsdk:"kcd_username"`

	LDAPAdmin    types.String `tfsdk:"ldap_admin"`
	CertCheckASI types.Bool   `tfsdk:"cert_check_asi"`
	CertCheckCN  types.Bool   `tfsdk:"cert_check_cn"`

	IdpEntityId  types.String `tfsdk:"idp_entity_id"`
	IdpSsoUrl    types.String `tfsdk:"idp_sso_url"`
	IdpLogoffUrl types.String `tfsdk:"idp_logoff_url"`
	IdpCert      types.String `tfsdk:"idp_cert"`
	IdpMatchCert types.Bool   `tfsdk:"idp_match_cert"`
	SpEntityId   types.String `tfsdk:"sp_entity_id"`
	SpCert       types.String `tfsdk:"sp_cert"`

	OidcAppId       types.String `tfsdk:"oidc_app_id"`
	OidcRedirectUri types.String `tfsdk:"oidc_redirect_uri"`
	OidcAuthEpUrl   types.String `tfsdk:"oidc_auth_ep_url"`
	OidcTokenEpUrl  types.String `tfsdk:"oidc_token_ep_url"`
	OidcLogoffUrl   types.String `tfsdk:"oidc_logoff_url"`
}

func (d *SSODomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sso_domain"
}

func (d *SSODomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Kemp LoadMaster SSO authentication domain by name.",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true},
			"name":     schema.StringAttribute{Required: true, MarkdownDescription: "Domain name to look up (e.g. `example.com`)."},
			"auth_type": schema.StringAttribute{Computed: true},
			"ldap_endpoint":              schema.StringAttribute{Computed: true},
			"ldap_endpoint_health_check": schema.BoolAttribute{Computed: true},
			"radius_send_nas_id":         schema.BoolAttribute{Computed: true},
			"radius_nas_id":              schema.StringAttribute{Computed: true},
			"logon_fmt":                  schema.StringAttribute{Computed: true},
			"logon_fmt2":                 schema.StringAttribute{Computed: true},
			"logon_domain":               schema.StringAttribute{Computed: true},
			"logon_transcode":            schema.BoolAttribute{Computed: true},
			"server":                     schema.StringAttribute{Computed: true},
			"user_acc_control":           schema.Int64Attribute{Computed: true},
			"max_failed_auths":           schema.Int64Attribute{Computed: true},
			"reset_fail_tout":            schema.Int64Attribute{Computed: true},
			"unblock_tout":               schema.Int64Attribute{Computed: true},
			"sess_tout_idle_pub":         schema.Int64Attribute{Computed: true},
			"sess_tout_duration_pub":     schema.Int64Attribute{Computed: true},
			"sess_tout_idle_priv":        schema.Int64Attribute{Computed: true},
			"sess_tout_duration_priv":    schema.Int64Attribute{Computed: true},
			"sess_tout_type":             schema.StringAttribute{Computed: true},
			"server_side":                schema.BoolAttribute{Computed: true},
			"kerberos_domain":            schema.StringAttribute{Computed: true},
			"kerberos_kdc":               schema.StringAttribute{Computed: true},
			"kcd_username":               schema.StringAttribute{Computed: true},
			"ldap_admin":                 schema.StringAttribute{Computed: true},
			"cert_check_asi":             schema.BoolAttribute{Computed: true},
			"cert_check_cn":              schema.BoolAttribute{Computed: true},
			"idp_entity_id":              schema.StringAttribute{Computed: true},
			"idp_sso_url":                schema.StringAttribute{Computed: true},
			"idp_logoff_url":             schema.StringAttribute{Computed: true},
			"idp_cert":                   schema.StringAttribute{Computed: true},
			"idp_match_cert":             schema.BoolAttribute{Computed: true},
			"sp_entity_id":               schema.StringAttribute{Computed: true},
			"sp_cert":                    schema.StringAttribute{Computed: true},
			"oidc_app_id":                schema.StringAttribute{Computed: true},
			"oidc_redirect_uri":          schema.StringAttribute{Computed: true},
			"oidc_auth_ep_url":           schema.StringAttribute{Computed: true},
			"oidc_token_ep_url":          schema.StringAttribute{Computed: true},
			"oidc_logoff_url":            schema.StringAttribute{Computed: true},
		},
	}
}

func (d *SSODomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*loadmaster.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configure type", fmt.Sprintf("Expected *loadmaster.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *SSODomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SSODomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dom, err := d.client.ShowSSODomain(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading SSO domain", err.Error())
		return
	}

	data.Id = types.StringValue(strings.ToLower(dom.Name))
	data.Name = types.StringValue(strings.ToLower(dom.Name))
	data.AuthType = types.StringValue(dom.AuthType)
	data.LDAPEndpoint = types.StringValue(dom.LDAPEndpoint)
	data.LDAPEpHC = types.BoolValue(dom.LDAPEpHC != 0)
	data.RadiusSendNasId = types.BoolValue(dom.RadiusSendNasId == "1")
	data.RadiusNasId = types.StringValue(dom.RadiusNasId)
	data.LogonFmt = types.StringValue(dom.LogonFmt)
	data.LogonFmt2 = types.StringValue(dom.LogonFmt2)
	data.LogonDomain = types.StringValue(dom.LogonDomain)
	data.LogonTranscode = types.BoolValue(dom.LogonTranscode != 0)
	data.UserAccControl = types.Int64Value(int64(dom.UserAccControl))
	data.MaxFailedAuths = types.Int64Value(int64(dom.MaxFailedAuths))
	data.ResetFailTout = types.Int64Value(int64(dom.ResetFailTout))
	data.UnblockTout = types.Int64Value(int64(dom.UnblockTout))
	data.SessToutIdlePub = types.Int64Value(int64(dom.SessToutIdlePub))
	data.SessToutDurPub = types.Int64Value(int64(dom.SessToutDurPub))
	data.SessToutIdlePriv = types.Int64Value(int64(dom.SessToutIdlePriv))
	data.SessToutDurPriv = types.Int64Value(int64(dom.SessToutDurPriv))
	data.SessToutType = types.StringValue(dom.SessToutType)
	data.ServerSide = types.BoolValue(dom.ServerSide == "1")
	data.KerberosDomain = types.StringValue(dom.KerberosDomain)
	data.KerberosKDC = types.StringValue(dom.KerberosKDC)
	data.KCDUsername = types.StringValue(dom.KCDUsername)
	data.LDAPAdmin = types.StringValue(dom.LDAPAdmin)
	data.CertCheckASI = types.BoolValue(dom.CertCheckASI != "" && dom.CertCheckASI != "Not Specified" && dom.CertCheckASI != "0")
	data.CertCheckCN = types.BoolValue(dom.CertCheckCN == "1")
	data.IdpEntityId = types.StringValue(dom.IdpEntityId)
	data.IdpSsoUrl = types.StringValue(dom.IdpSsoUrl)
	data.IdpLogoffUrl = types.StringValue(dom.IdpLogoffUrl)
	data.IdpCert = types.StringValue(dom.IdpCert)
	data.IdpMatchCert = types.BoolValue(dom.IdpMatchCert == "1")
	data.SpEntityId = types.StringValue(dom.SpEntityId)
	data.SpCert = types.StringValue(dom.SpCert)
	data.OidcAppId = types.StringValue(dom.OidcAppId)
	data.OidcRedirectUri = types.StringValue(dom.OidcRedirectUri)
	data.OidcAuthEpUrl = types.StringValue(dom.OidcAuthEpUrl)
	data.OidcTokenEpUrl = types.StringValue(dom.OidcTokenEpUrl)
	data.OidcLogoffUrl = types.StringValue(dom.OidcLogoffUrl)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

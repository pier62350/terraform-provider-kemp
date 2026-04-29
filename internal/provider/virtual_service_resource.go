// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &VirtualServiceResource{}
	_ resource.ResourceWithImportState = &VirtualServiceResource{}
)

func NewVirtualServiceResource() resource.Resource {
	return &VirtualServiceResource{}
}

type VirtualServiceResource struct {
	client *loadmaster.Client
}

type VirtualServiceResourceModel struct {
	Id              types.String `tfsdk:"id"`
	Address         types.String `tfsdk:"address"`
	Port            types.String `tfsdk:"port"`
	Protocol        types.String `tfsdk:"protocol"`
	Type            types.String `tfsdk:"type"`
	Nickname        types.String `tfsdk:"nickname"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	SSLAcceleration types.Bool   `tfsdk:"ssl_acceleration"`
	CertFiles       types.List   `tfsdk:"cert_files"`

	// Standard options
	Schedule            types.String `tfsdk:"schedule"`
	PersistTimeout      types.String `tfsdk:"persist_timeout"`
	Idletime            types.Int64  `tfsdk:"idletime"`
	ForceL7             types.Bool   `tfsdk:"force_l7"`
	ForceL4             types.Bool   `tfsdk:"force_l4"`
	Transparent         types.Bool   `tfsdk:"transparent"`
	UseForSnat          types.Bool   `tfsdk:"use_for_snat"`
	Cache               types.Bool   `tfsdk:"cache"`
	Compress            types.Bool   `tfsdk:"compress"`
	AllowHTTP2          types.Bool   `tfsdk:"allow_http2"`
	SSLReverse          types.Bool   `tfsdk:"ssl_reverse"`
	SSLReencrypt        types.Bool   `tfsdk:"ssl_reencrypt"`
	AddVia              types.String `tfsdk:"add_via"`
	RefreshPersist      types.Bool   `tfsdk:"refresh_persist"`
	RsMinimum           types.Int64  `tfsdk:"rs_minimum"`
	Bandwidth           types.Int64  `tfsdk:"bandwidth"`
	ConnsPerSecLimit    types.Int64  `tfsdk:"conns_per_sec_limit"`
	RequestsPerSecLimit types.Int64  `tfsdk:"requests_per_sec_limit"`
	MaxConnsLimit       types.Int64  `tfsdk:"max_conns_limit"`

	// Health checks
	CheckType            types.String `tfsdk:"check_type"`
	CheckPort            types.String `tfsdk:"check_port"`
	ChkInterval          types.Int64  `tfsdk:"chk_interval"`
	ChkTimeout           types.Int64  `tfsdk:"chk_timeout"`
	ChkRetryCount        types.Int64  `tfsdk:"chk_retry_count"`
	NeedHostName         types.Bool   `tfsdk:"need_host_name"`
	CheckUseHTTP11       types.Bool   `tfsdk:"check_use_http11"`
	CheckUseGet          types.String `tfsdk:"check_use_get"`
	MatchLen             types.Int64  `tfsdk:"match_len"`
	EnhancedHealthChecks types.Bool   `tfsdk:"enhanced_health_checks"`

	// ESP
	EspEnabled             types.Bool   `tfsdk:"esp_enabled"`
	EspAllowedHosts        types.String `tfsdk:"esp_allowed_hosts"`
	EspAllowedDirectories  types.String `tfsdk:"esp_allowed_directories"`
	EspInputAuthMode       types.String `tfsdk:"esp_input_auth_mode"`
	EspOutputAuthMode      types.String `tfsdk:"esp_output_auth_mode"`
	EspIncludeNestedGroups types.Bool   `tfsdk:"esp_include_nested_groups"`
	EspDisplayPubPriv      types.Bool   `tfsdk:"esp_display_pub_priv"`
	EspLogs                types.Bool   `tfsdk:"esp_logs"`

	// WAF
	WafInterceptMode         types.String `tfsdk:"waf_intercept_mode"`
	WafBlockingParanoia      types.Int64  `tfsdk:"waf_blocking_paranoia"`
	WafAlertThreshold        types.Int64  `tfsdk:"waf_alert_threshold"`
	WafIpReputationBlocking  types.Bool   `tfsdk:"waf_ip_reputation_blocking"`
}

func (r *VirtualServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_service"
}

func (r *VirtualServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kemp LoadMaster virtual service (VS).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "LoadMaster `Index` of the virtual service.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "IP address of an interface attached to the LoadMaster.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Listening port of the virtual service.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Layer-4 protocol: `tcp` or `udp`.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "VS type — one of `gen`, `http`, `http2`, `ts`, `tls`, `log`.",
				Optional:            true,
				Computed:            true,
			},
			"nickname": schema.StringAttribute{
				MarkdownDescription: "Friendly name for the virtual service.",
				Optional:            true,
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the virtual service is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"ssl_acceleration": schema.BoolAttribute{
				MarkdownDescription: "Enable SSL/TLS termination on the LoadMaster (a.k.a. SSL acceleration). Requires `cert_files` to be set.",
				Optional:            true,
				Computed:            true,
			},
			"cert_files": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Names of certificates (as stored on the LoadMaster) attached to this virtual service. Multiple entries enable SNI: LoadMaster picks the cert whose subject matches the client's TLS SNI hostname. Order matters — the first cert is the default.",
				Optional:            true,
				Computed:            true,
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "Load-balancing algorithm: `rr` (round-robin), `wlc` (weighted least-connections), `lc` (least-connections), `pi` (proximity IP), `ph` (persistent hash), etc.",
				Optional:            true,
				Computed:            true,
			},
			"persist_timeout": schema.StringAttribute{
				MarkdownDescription: "Persistence timeout in seconds. `0` disables persistence.",
				Optional:            true,
				Computed:            true,
			},
			"idletime": schema.Int64Attribute{
				MarkdownDescription: "Idle connection timeout in seconds. Default is 660.",
				Optional:            true,
				Computed:            true,
			},
			"force_l7": schema.BoolAttribute{
				MarkdownDescription: "Force Layer-7 processing even when the VS is configured as Layer-4.",
				Optional:            true,
				Computed:            true,
			},
			"force_l4": schema.BoolAttribute{
				MarkdownDescription: "Force Layer-4 processing (bypass Layer-7 inspection).",
				Optional:            true,
				Computed:            true,
			},
			"transparent": schema.BoolAttribute{
				MarkdownDescription: "Transparent mode — preserves the client IP address when forwarding to real servers.",
				Optional:            true,
				Computed:            true,
			},
			"use_for_snat": schema.BoolAttribute{
				MarkdownDescription: "Use this VS as the source NAT address for outbound connections.",
				Optional:            true,
				Computed:            true,
			},
			"cache": schema.BoolAttribute{
				MarkdownDescription: "Enable HTTP response caching on the LoadMaster for this VS.",
				Optional:            true,
				Computed:            true,
			},
			"compress": schema.BoolAttribute{
				MarkdownDescription: "Enable HTTP response compression (gzip) for this VS.",
				Optional:            true,
				Computed:            true,
			},
			"allow_http2": schema.BoolAttribute{
				MarkdownDescription: "Enable HTTP/2 support on this VS.",
				Optional:            true,
				Computed:            true,
			},
			"ssl_reverse": schema.BoolAttribute{
				MarkdownDescription: "Re-encrypt connections to real servers using SSL (SSL offload in reverse).",
				Optional:            true,
				Computed:            true,
			},
			"ssl_reencrypt": schema.BoolAttribute{
				MarkdownDescription: "Re-encrypt to real servers using the same SSL session parameters as the client.",
				Optional:            true,
				Computed:            true,
			},
			"add_via": schema.StringAttribute{
				MarkdownDescription: "Whether to add a `Via` header to proxied requests: `no`, `add`, or `replace`.",
				Optional:            true,
				Computed:            true,
			},
			"refresh_persist": schema.BoolAttribute{
				MarkdownDescription: "Refresh the persistence entry on every request (not just the first).",
				Optional:            true,
				Computed:            true,
			},
			"rs_minimum": schema.Int64Attribute{
				MarkdownDescription: "Minimum number of active real servers required before the VS is marked up. `0` means no minimum.",
				Optional:            true,
				Computed:            true,
			},
			"check_type": schema.StringAttribute{
				MarkdownDescription: "Health check type: `tcp`, `http`, `https`, `icmp`, `smtp`, `nntp`, `ftp`, `dns`, `pop3`, `imap`, `rdp`, `snmp`, `ldap`, `none`, etc.",
				Optional:            true,
				Computed:            true,
			},
			"check_port": schema.StringAttribute{
				MarkdownDescription: "Port used for health checks. `0` means use the VS port.",
				Optional:            true,
				Computed:            true,
			},
			"chk_interval": schema.Int64Attribute{
				MarkdownDescription: "Interval between health checks in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"chk_timeout": schema.Int64Attribute{
				MarkdownDescription: "Health check timeout in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"chk_retry_count": schema.Int64Attribute{
				MarkdownDescription: "Number of consecutive failed health checks before a real server is marked down.",
				Optional:            true,
				Computed:            true,
			},
			"need_host_name": schema.BoolAttribute{
				MarkdownDescription: "Send the VS hostname in the HTTP `Host` header during health checks.",
				Optional:            true,
				Computed:            true,
			},
			"check_use_http11": schema.BoolAttribute{
				MarkdownDescription: "Use HTTP/1.1 for HTTP-based health checks (instead of HTTP/1.0).",
				Optional:            true,
				Computed:            true,
			},
			"check_use_get": schema.StringAttribute{
				MarkdownDescription: "HTTP method used for health checks: `head` (default) or `get`.",
				Optional:            true,
				Computed:            true,
			},
			"match_len": schema.Int64Attribute{
				MarkdownDescription: "Number of bytes of the health check response body to inspect for the match pattern.",
				Optional:            true,
				Computed:            true,
			},
			"enhanced_health_checks": schema.BoolAttribute{
				MarkdownDescription: "Enable enhanced health checks (sends a more complete HTTP request).",
				Optional:            true,
				Computed:            true,
			},
			"bandwidth": schema.Int64Attribute{
				MarkdownDescription: "Bandwidth limit in Mbps. `0` means unlimited.",
				Optional:            true,
				Computed:            true,
			},
			"conns_per_sec_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum new connections per second. `0` means unlimited.",
				Optional:            true,
				Computed:            true,
			},
			"requests_per_sec_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum HTTP requests per second. `0` means unlimited.",
				Optional:            true,
				Computed:            true,
			},
			"max_conns_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum concurrent connections. `0` means unlimited.",
				Optional:            true,
				Computed:            true,
			},
			"esp_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable Kemp Edge Security Pack (ESP) on this VS — pre-auth, SSO, header injection, etc. Requires `type = http` and typically `ssl_acceleration = true`.",
				Optional:            true,
				Computed:            true,
			},
			"esp_allowed_hosts": schema.StringAttribute{
				MarkdownDescription: "Newline-separated list of hostnames the VS will accept for ESP. Empty matches all.",
				Optional:            true,
				Computed:            true,
			},
			"esp_allowed_directories": schema.StringAttribute{
				MarkdownDescription: "Newline-separated list of allowed URI prefixes when ESP is on.",
				Optional:            true,
				Computed:            true,
			},
			"esp_input_auth_mode": schema.StringAttribute{
				MarkdownDescription: "Client-side authentication mode. Known values: `none`, `basic`, `form`. Other numeric values from LoadMaster docs are also accepted.",
				Optional:            true,
				Computed:            true,
			},
			"esp_output_auth_mode": schema.StringAttribute{
				MarkdownDescription: "Server-side authentication mode for the upstream. Known values: `none`, `basic`, `form`, `kcd` (Kerberos Constrained Delegation). Other numeric values from LoadMaster docs are also accepted.",
				Optional:            true,
				Computed:            true,
			},
			"esp_include_nested_groups": schema.BoolAttribute{
				MarkdownDescription: "When ESP authorizes against AD groups, follow nested-group memberships.",
				Optional:            true,
				Computed:            true,
			},
			"esp_display_pub_priv": schema.BoolAttribute{
				MarkdownDescription: "Display the public/private toggle on the ESP login form.",
				Optional:            true,
				Computed:            true,
			},
			"esp_logs": schema.BoolAttribute{
				MarkdownDescription: "Enable extended ESP logging for this VS.",
				Optional:            true,
				Computed:            true,
			},
			"waf_intercept_mode": schema.StringAttribute{
				MarkdownDescription: "WAF intercept mode: `disabled`, `legacy` (Legacy WAF), or `owasp` (OWASP WAF).",
				Optional:            true,
				Computed:            true,
			},
			"waf_blocking_paranoia": schema.Int64Attribute{
				MarkdownDescription: "OWASP paranoia level (0-4). Higher = more rules trigger, more false positives.",
				Optional:            true,
				Computed:            true,
			},
			"waf_alert_threshold": schema.Int64Attribute{
				MarkdownDescription: "Anomaly score that triggers blocking. Set to 0 for detection-only (audit) mode.",
				Optional:            true,
				Computed:            true,
			},
			"waf_ip_reputation_blocking": schema.BoolAttribute{
				MarkdownDescription: "Enable IP Reputation Blocking. When enabled, requests from IP addresses with a bad reputation are blocked by the WAF.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *VirtualServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*loadmaster.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *loadmaster.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *VirtualServiceResource) paramsFromModel(ctx context.Context, m VirtualServiceResourceModel) (loadmaster.VirtualServiceParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	p := loadmaster.VirtualServiceParams{
		NickName: m.Nickname.ValueString(),
		VSType:   m.Type.ValueString(),
	}
	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		p.Enable = boolPtr(m.Enabled.ValueBool())
	}
	if !m.SSLAcceleration.IsNull() && !m.SSLAcceleration.IsUnknown() {
		p.SSLAcceleration = boolPtr(m.SSLAcceleration.ValueBool())
	}
	if !m.CertFiles.IsNull() && !m.CertFiles.IsUnknown() {
		var certs []string
		diags.Append(m.CertFiles.ElementsAs(ctx, &certs, false)...)
		if !diags.HasError() {
			p.CertFile = strings.Join(certs, ",")
		}
	}

	if !m.Schedule.IsNull() && !m.Schedule.IsUnknown() {
		p.Schedule = m.Schedule.ValueString()
	}
	if !m.PersistTimeout.IsNull() && !m.PersistTimeout.IsUnknown() {
		p.PersistTimeout = m.PersistTimeout.ValueString()
	}
	if !m.Idletime.IsNull() && !m.Idletime.IsUnknown() {
		v := int32(m.Idletime.ValueInt64())
		p.Idletime = &v
	}
	if !m.ForceL7.IsNull() && !m.ForceL7.IsUnknown() {
		p.ForceL7 = boolPtr(m.ForceL7.ValueBool())
	}
	if !m.ForceL4.IsNull() && !m.ForceL4.IsUnknown() {
		p.ForceL4 = boolPtr(m.ForceL4.ValueBool())
	}
	if !m.Transparent.IsNull() && !m.Transparent.IsUnknown() {
		p.Transparent = boolPtr(m.Transparent.ValueBool())
	}
	if !m.UseForSnat.IsNull() && !m.UseForSnat.IsUnknown() {
		p.UseforSnat = boolPtr(m.UseForSnat.ValueBool())
	}
	if !m.Cache.IsNull() && !m.Cache.IsUnknown() {
		p.Cache = boolPtr(m.Cache.ValueBool())
	}
	if !m.Compress.IsNull() && !m.Compress.IsUnknown() {
		p.Compress = boolPtr(m.Compress.ValueBool())
	}
	if !m.AllowHTTP2.IsNull() && !m.AllowHTTP2.IsUnknown() {
		p.AllowHTTP2 = boolPtr(m.AllowHTTP2.ValueBool())
	}
	if !m.SSLReverse.IsNull() && !m.SSLReverse.IsUnknown() {
		p.SSLReverse = boolPtr(m.SSLReverse.ValueBool())
	}
	if !m.SSLReencrypt.IsNull() && !m.SSLReencrypt.IsUnknown() {
		p.SSLReencrypt = boolPtr(m.SSLReencrypt.ValueBool())
	}
	if !m.AddVia.IsNull() && !m.AddVia.IsUnknown() {
		p.AddVia = addViaToAPI(m.AddVia.ValueString())
	}
	if !m.RefreshPersist.IsNull() && !m.RefreshPersist.IsUnknown() {
		p.RefreshPersist = boolPtr(m.RefreshPersist.ValueBool())
	}
	if !m.RsMinimum.IsNull() && !m.RsMinimum.IsUnknown() {
		v := int32(m.RsMinimum.ValueInt64())
		p.RsMinimum = &v
	}
	if !m.CheckType.IsNull() && !m.CheckType.IsUnknown() {
		p.CheckType = m.CheckType.ValueString()
	}
	if !m.CheckPort.IsNull() && !m.CheckPort.IsUnknown() {
		p.CheckPort = m.CheckPort.ValueString()
	}
	if !m.ChkInterval.IsNull() && !m.ChkInterval.IsUnknown() {
		v := int32(m.ChkInterval.ValueInt64())
		p.ChkInterval = &v
	}
	if !m.ChkTimeout.IsNull() && !m.ChkTimeout.IsUnknown() {
		v := int32(m.ChkTimeout.ValueInt64())
		p.ChkTimeout = &v
	}
	if !m.ChkRetryCount.IsNull() && !m.ChkRetryCount.IsUnknown() {
		v := int32(m.ChkRetryCount.ValueInt64())
		p.ChkRetryCount = &v
	}
	if !m.NeedHostName.IsNull() && !m.NeedHostName.IsUnknown() {
		p.NeedHostName = boolPtr(m.NeedHostName.ValueBool())
	}
	if !m.CheckUseHTTP11.IsNull() && !m.CheckUseHTTP11.IsUnknown() {
		p.CheckUseHTTP11 = boolPtr(m.CheckUseHTTP11.ValueBool())
	}
	if !m.CheckUseGet.IsNull() && !m.CheckUseGet.IsUnknown() {
		p.CheckUseGet = checkUseGetToAPI(m.CheckUseGet.ValueString())
	}
	if !m.MatchLen.IsNull() && !m.MatchLen.IsUnknown() {
		v := int32(m.MatchLen.ValueInt64())
		p.MatchLen = &v
	}
	if !m.EnhancedHealthChecks.IsNull() && !m.EnhancedHealthChecks.IsUnknown() {
		p.EnhancedHealthChecks = boolPtr(m.EnhancedHealthChecks.ValueBool())
	}
	if !m.Bandwidth.IsNull() && !m.Bandwidth.IsUnknown() {
		v := int32(m.Bandwidth.ValueInt64())
		p.Bandwidth = &v
	}
	if !m.ConnsPerSecLimit.IsNull() && !m.ConnsPerSecLimit.IsUnknown() {
		v := int32(m.ConnsPerSecLimit.ValueInt64())
		p.ConnsPerSecLimit = &v
	}
	if !m.RequestsPerSecLimit.IsNull() && !m.RequestsPerSecLimit.IsUnknown() {
		v := int32(m.RequestsPerSecLimit.ValueInt64())
		p.RequestsPerSecLimit = &v
	}
	if !m.MaxConnsLimit.IsNull() && !m.MaxConnsLimit.IsUnknown() {
		v := int32(m.MaxConnsLimit.ValueInt64())
		p.MaxConnsLimit = &v
	}

	if !m.EspEnabled.IsNull() && !m.EspEnabled.IsUnknown() {
		p.EspEnabled = boolPtr(m.EspEnabled.ValueBool())
	}
	if !m.EspAllowedHosts.IsNull() && !m.EspAllowedHosts.IsUnknown() {
		p.AllowedHosts = m.EspAllowedHosts.ValueString()
	}
	if !m.EspAllowedDirectories.IsNull() && !m.EspAllowedDirectories.IsUnknown() {
		p.AllowedDirectories = m.EspAllowedDirectories.ValueString()
	}
	if !m.EspInputAuthMode.IsNull() && !m.EspInputAuthMode.IsUnknown() {
		p.InputAuthMode = espInputAuthModeToAPI(m.EspInputAuthMode.ValueString())
	}
	if !m.EspOutputAuthMode.IsNull() && !m.EspOutputAuthMode.IsUnknown() {
		p.OutputAuthMode = espOutputAuthModeToAPI(m.EspOutputAuthMode.ValueString())
	}
	if !m.EspIncludeNestedGroups.IsNull() && !m.EspIncludeNestedGroups.IsUnknown() {
		p.IncludeNestedGroups = boolPtr(m.EspIncludeNestedGroups.ValueBool())
	}
	if !m.EspDisplayPubPriv.IsNull() && !m.EspDisplayPubPriv.IsUnknown() {
		p.DisplayPubPriv = boolPtr(m.EspDisplayPubPriv.ValueBool())
	}
	if !m.EspLogs.IsNull() && !m.EspLogs.IsUnknown() {
		p.EspLogs = boolPtr(m.EspLogs.ValueBool())
	}

	if !m.WafInterceptMode.IsNull() && !m.WafInterceptMode.IsUnknown() {
		p.InterceptMode = wafInterceptModeToAPI(m.WafInterceptMode.ValueString())
	}
	if !m.WafBlockingParanoia.IsNull() && !m.WafBlockingParanoia.IsUnknown() {
		v := int32(m.WafBlockingParanoia.ValueInt64())
		p.BlockingParanoia = &v
	}
	if !m.WafAlertThreshold.IsNull() && !m.WafAlertThreshold.IsUnknown() {
		v := int32(m.WafAlertThreshold.ValueInt64())
		p.AlertThreshold = &v
	}
	if !m.WafIpReputationBlocking.IsNull() && !m.WafIpReputationBlocking.IsUnknown() {
		p.IPReputationBlocking = boolPtr(m.WafIpReputationBlocking.ValueBool())
	}
	return p, diags
}

func (r *VirtualServiceResource) writeState(ctx context.Context, vs *loadmaster.VirtualService, m *VirtualServiceResourceModel) diag.Diagnostics {
	m.Id = types.StringValue(strconv.Itoa(int(vs.Index)))
	m.Address = types.StringValue(vs.Address)
	m.Port = types.StringValue(vs.Port)
	m.Protocol = types.StringValue(vs.Protocol)
	m.Type = types.StringValue(vs.VSType)
	m.Nickname = types.StringValue(vs.NickName)
	if vs.Enable != nil {
		m.Enabled = types.BoolValue(*vs.Enable)
	} else {
		m.Enabled = types.BoolValue(false)
	}
	if vs.SSLAcceleration != nil {
		m.SSLAcceleration = types.BoolValue(*vs.SSLAcceleration)
	} else {
		m.SSLAcceleration = types.BoolValue(false)
	}

	var certs []string
	if vs.CertFile != "" {
		certs = strings.Split(vs.CertFile, ",")
		for i := range certs {
			certs[i] = strings.TrimSpace(certs[i])
		}
	}
	listVal, diags := types.ListValueFrom(ctx, types.StringType, certs)
	m.CertFiles = listVal

	m.Schedule = types.StringValue(vs.Schedule)
	m.PersistTimeout = types.StringValue(vs.PersistTimeout)
	m.Idletime = int64FromPtr(vs.Idletime)
	m.ForceL7 = boolFromPtr(vs.ForceL7)
	m.ForceL4 = boolFromPtr(vs.ForceL4)
	m.Transparent = boolFromPtr(vs.Transparent)
	m.UseForSnat = boolFromPtr(vs.UseforSnat)
	m.Cache = boolFromPtr(vs.Cache)
	m.Compress = boolFromPtr(vs.Compress)
	m.AllowHTTP2 = boolFromPtr(vs.AllowHTTP2)
	m.SSLReverse = boolFromPtr(vs.SSLReverse)
	m.SSLReencrypt = boolFromPtr(vs.SSLReencrypt)
	m.AddVia = types.StringValue(addViaFromAPI(vs.AddVia))
	m.RefreshPersist = boolFromPtr(vs.RefreshPersist)
	m.RsMinimum = int64FromPtr(vs.RsMinimum)
	m.CheckType = types.StringValue(vs.CheckType)
	m.CheckPort = types.StringValue(vs.CheckPort)
	m.ChkInterval = int64FromPtr(vs.ChkInterval)
	m.ChkTimeout = int64FromPtr(vs.ChkTimeout)
	m.ChkRetryCount = int64FromPtr(vs.ChkRetryCount)
	m.NeedHostName = boolFromPtr(vs.NeedHostName)
	m.CheckUseHTTP11 = boolFromPtr(vs.CheckUseHTTP11)
	m.CheckUseGet = types.StringValue(checkUseGetFromAPI(vs.CheckUseGet))
	m.MatchLen = int64FromPtr(vs.MatchLen)
	m.EnhancedHealthChecks = boolFromPtr(vs.EnhancedHealthChecks)
	m.Bandwidth = int64FromPtr(vs.Bandwidth)
	m.ConnsPerSecLimit = int64FromPtr(vs.ConnsPerSecLimit)
	m.RequestsPerSecLimit = int64FromPtr(vs.RequestsPerSecLimit)
	m.MaxConnsLimit = int64FromPtr(vs.MaxConnsLimit)

	m.EspEnabled = boolFromPtr(vs.EspEnabled)
	m.EspAllowedHosts = types.StringValue(vs.AllowedHosts)
	m.EspAllowedDirectories = types.StringValue(vs.AllowedDirectories)
	m.EspInputAuthMode = types.StringValue(espInputAuthModeFromAPI(vs.InputAuthMode))
	m.EspOutputAuthMode = types.StringValue(espOutputAuthModeFromAPI(vs.OutputAuthMode))
	m.EspIncludeNestedGroups = boolFromPtr(vs.IncludeNestedGroups)
	m.EspDisplayPubPriv = boolFromPtr(vs.DisplayPubPriv)
	m.EspLogs = boolFromPtr(vs.EspLogs)

	m.WafInterceptMode = types.StringValue(wafInterceptModeFromAPI(vs.InterceptMode))
	m.WafBlockingParanoia = int64FromPtr(vs.BlockingParanoia)
	m.WafAlertThreshold = int64FromPtr(vs.AlertThreshold)
	m.WafIpReputationBlocking = boolFromPtr(vs.IPReputationBlocking)

	return diags
}

func (r *VirtualServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params, d := r.paramsFromModel(ctx, data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.AddVirtualService(ctx, data.Address.ValueString(), data.Port.ValueString(), data.Protocol.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating virtual service", err.Error())
		return
	}
	tflog.Trace(ctx, "created virtual service", map[string]any{"index": vs.Index})

	resp.Diagnostics.Append(r.writeState(ctx, vs, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.ShowVirtualService(ctx, data.Id.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading virtual service", err.Error())
		return
	}

	resp.Diagnostics.Append(r.writeState(ctx, vs, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params, d := r.paramsFromModel(ctx, data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.ModifyVirtualService(ctx, data.Id.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating virtual service", err.Error())
		return
	}

	resp.Diagnostics.Append(r.writeState(ctx, vs, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVirtualService(ctx, data.Id.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting virtual service", err.Error())
	}
}

func (r *VirtualServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	vs, err := r.client.ShowVirtualService(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing virtual service", err.Error())
		return
	}

	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(r.writeState(ctx, vs, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

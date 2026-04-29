// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &SubVirtualServiceResource{}
	_ resource.ResourceWithImportState = &SubVirtualServiceResource{}
)

func NewSubVirtualServiceResource() resource.Resource { return &SubVirtualServiceResource{} }

type SubVirtualServiceResource struct {
	client *loadmaster.Client
}

type SubVirtualServiceResourceModel struct {
	Id              types.String `tfsdk:"id"`
	ParentId        types.String `tfsdk:"parent_id"`
	Nickname        types.String `tfsdk:"nickname"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Type            types.String `tfsdk:"type"`
	SSLAcceleration types.Bool   `tfsdk:"ssl_acceleration"`
	CertFiles       types.List   `tfsdk:"cert_files"`

	// Standard options
	Schedule            types.String `tfsdk:"schedule"`
	Persist             types.String `tfsdk:"persist"`
	PersistTimeout      types.String `tfsdk:"persist_timeout"`
	Idletime            types.Int64  `tfsdk:"idletime"`
	ServerInit          types.Int64  `tfsdk:"server_init"`
	ForceL7             types.Bool   `tfsdk:"force_l7"`
	ForceL4             types.Bool   `tfsdk:"force_l4"`
	Transparent         types.Bool   `tfsdk:"transparent"`
	UseForSnat          types.Bool   `tfsdk:"use_for_snat"`
	MultiConnect        types.Bool   `tfsdk:"multi_connect"`
	Cache               types.Bool   `tfsdk:"cache"`
	Compress            types.Bool   `tfsdk:"compress"`
	AllowHTTP2          types.Bool   `tfsdk:"allow_http2"`
	SSLReverse          types.Bool   `tfsdk:"ssl_reverse"`
	SSLReencrypt        types.Bool   `tfsdk:"ssl_reencrypt"`
	PassSni             types.Bool   `tfsdk:"pass_sni"`
	PassCipher          types.Bool   `tfsdk:"pass_cipher"`
	Verify              types.Int64  `tfsdk:"verify"`
	ClientCert          types.Int64  `tfsdk:"client_cert"`
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
	WafInterceptMode        types.String `tfsdk:"waf_intercept_mode"`
	WafBlockingParanoia     types.Int64  `tfsdk:"waf_blocking_paranoia"`
	WafAlertThreshold       types.Int64  `tfsdk:"waf_alert_threshold"`
	WafIpReputationBlocking types.Bool   `tfsdk:"waf_ip_reputation_blocking"`
}

func (r *SubVirtualServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sub_virtual_service"
}

func (r *SubVirtualServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Sub-Virtual Service (SubVS) under a parent ` + "`kemp_virtual_service`" + `.

A SubVS shares the parent's listening address/port/protocol and adds L7 routing on top — typically used to dispatch traffic by host/path to different real-server pools. Sub-VS creation goes through the parent's ` + "`modvs`" + ` with ` + "`createsubvs`" + `; thereafter the SubVS has its own Index used for CRUD.

The SubVS exposes the same SSL + ESP + WAF surface as the parent ` + "`kemp_virtual_service`" + ` resource — a SubVS can carry its own auth config and WAF posture independently of its siblings.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "LoadMaster `Index` assigned to the SubVS.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"parent_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "`Index` of the parent virtual service this SubVS attaches to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"nickname": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Friendly name for the SubVS shown in the WUI."},
			"enabled":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Whether the SubVS is enabled. Default: `true`."},
			"type":     schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. VS type — `gen`, `http`, `http2`, `ts`, `tls`, `log`. Default: `gen`."},
			"ssl_acceleration": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Enable SSL/TLS termination on this SubVS. Requires `cert_files` to be set. Default: `false`.",
			},
			"cert_files": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Names of certificates attached to this SubVS (multiple entries enable SNI — LoadMaster picks by client SNI; first entry is the fallback). Default: empty.",
			},
			"schedule":         schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Load-balancing algorithm: `rr` (round-robin), `wlc` (weighted least-connections), `lc` (least-connections), `pi` (proximity IP), `ph` (persistent hash), etc. Default: `rr`."},
			"persist":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Persistence mode: `src` (source IP), `cookie`, `active-cookie`, `active-cookie-insert`, `ssl`, `sip`, `rdp`, `super`, `none`. Default: `none`. **Note:** LoadMaster does not return this field on read — stored in state as-set."},
			"persist_timeout":  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Persistence timeout in seconds. Default: `0` (persistence disabled)."},
			"idletime":         schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Idle connection timeout in seconds. Default: `660`."},
			"server_init":      schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Server-side connection initialisation timeout in seconds. Default: `0` (uses global setting)."},
			"force_l7":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Force Layer-7 processing. Default: `true` for `http`/`http2` types."},
			"force_l4":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Force Layer-4 processing (bypass Layer-7 inspection). Default: `false`."},
			"transparent":    schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Transparent mode — preserves the client IP address when forwarding to real servers. Default: `false`."},
			"use_for_snat":    schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Use this SubVS as the source NAT address for outbound connections. Default: `false`."},
			"multi_connect":   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Allow multiple simultaneous connections from the same client. Default: `false`."},
			"cache":          schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Enable HTTP response caching on the LoadMaster for this SubVS. Default: `false`."},
			"compress":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Enable HTTP response compression (gzip) for this SubVS. Default: `false`."},
			"allow_http2":    schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Enable HTTP/2 support on this SubVS. Default: `false`."},
			"ssl_reverse":    schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Re-encrypt connections to real servers using SSL. Default: `false`."},
			"ssl_reencrypt":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Re-encrypt to real servers using the same SSL session parameters as the client. Default: `false`."},
			"pass_sni":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Pass the TLS SNI hostname to real servers. Default: `false`."},
			"pass_cipher":    schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Pass the negotiated cipher suite to real servers. Default: `false`."},
			"verify":         schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Client certificate verification level: `0` = off (default), `1` = optional, `2` = mandatory, `3` = skip CA check."},
			"client_cert":    schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Client certificate forwarding: `0` = do not forward (default), `1` = forward if present, `2` = always require and forward."},
			"add_via":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Whether to add a `Via` header to proxied requests: `no` (default), `add`, or `replace`."},
			"refresh_persist": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Refresh the persistence entry on every request, not just the first. Default: `false`."},
			"rs_minimum":     schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Minimum number of active real servers required before the SubVS is marked up. Default: `0` (no minimum)."},
			"bandwidth":               schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Bandwidth limit in Mbps. Default: `0` (unlimited)."},
			"conns_per_sec_limit":     schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Maximum new connections per second. Default: `0` (unlimited)."},
			"requests_per_sec_limit":  schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Maximum HTTP requests per second. Default: `0` (unlimited)."},
			"max_conns_limit":         schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Maximum concurrent connections. Default: `0` (unlimited)."},
			"check_type":              schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Health check type: `tcp`, `http`, `https`, `icmp`, `smtp`, `nntp`, `ftp`, `dns`, `none`, etc. Default: `tcp`."},
			"check_port":              schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Port used for health checks. Default: `0` (use the SubVS listening port)."},
			"chk_interval":            schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Interval between health checks in seconds. Default: `0` (uses the global interval)."},
			"chk_timeout":             schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Health check timeout in seconds. Default: `0` (uses the global timeout)."},
			"chk_retry_count":         schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Consecutive failed checks before marking a real server down. Default: `0` (uses the global retry count)."},
			"need_host_name":          schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Send the VS hostname in the HTTP `Host` header during health checks. Default: `false`."},
			"check_use_http11":        schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Use HTTP/1.1 for HTTP-based health checks. Default: `false`."},
			"check_use_get":           schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. HTTP method for health checks: `head` (default) or `get`."},
			"match_len":               schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Bytes of the health check response body to inspect for a match pattern. Default: `0` (disabled)."},
			"enhanced_health_checks":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Enable enhanced health checks (sends a more complete HTTP request including headers). Default: `false`."},
			"esp_enabled":               schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Enable Kemp Edge Security Pack (ESP) on this SubVS. Default: `false`."},
			"esp_allowed_hosts":         schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Newline-separated list of hostnames the SubVS will accept for ESP. Empty string matches all hosts."},
			"esp_allowed_directories":   schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Newline-separated list of URI prefixes allowed through ESP. Empty string allows all paths."},
			"esp_input_auth_mode":       schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Client-side authentication mode: `none` (default), `basic`, or `form`."},
			"esp_output_auth_mode":      schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Server-side (upstream) authentication mode: `none` (default), `basic`, `form`, or `kcd` (Kerberos Constrained Delegation)."},
			"esp_include_nested_groups": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Follow nested AD group memberships when ESP authorizes users. Default: `false`."},
			"esp_display_pub_priv":      schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Display the public/private session toggle on the ESP login form. Default: `false`."},
			"esp_logs":                  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Enable extended ESP logging for this SubVS. Default: `false`."},
			"waf_intercept_mode":           schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. WAF intercept mode: `disabled` (default), `legacy` (Legacy WAF), or `owasp` (OWASP/ModSecurity WAF)."},
			"waf_blocking_paranoia":        schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. OWASP paranoia level (`0`–`4`). Higher values activate more rules. Default: `0`."},
			"waf_alert_threshold":          schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Anomaly score threshold that triggers blocking. Default: `0` (detection-only / audit mode)."},
			"waf_ip_reputation_blocking":   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Block requests from IP addresses with a bad reputation using the WAF IP Reputation database. Default: `false`."},
		},
	}
}

func (r *SubVirtualServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubVirtualServiceResource) paramsFromModel(ctx context.Context, m SubVirtualServiceResourceModel) (loadmaster.VirtualServiceParams, diag.Diagnostics) {
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
	if !m.Persist.IsNull() && !m.Persist.IsUnknown() {
		p.Persist = m.Persist.ValueString()
	}
	if !m.PersistTimeout.IsNull() && !m.PersistTimeout.IsUnknown() {
		p.PersistTimeout = m.PersistTimeout.ValueString()
	}
	if !m.Idletime.IsNull() && !m.Idletime.IsUnknown() {
		v := int32(m.Idletime.ValueInt64())
		p.Idletime = &v
	}
	if !m.ServerInit.IsNull() && !m.ServerInit.IsUnknown() {
		v := int32(m.ServerInit.ValueInt64())
		p.ServerInit = &v
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
	if !m.MultiConnect.IsNull() && !m.MultiConnect.IsUnknown() {
		p.MultiConnect = boolPtr(m.MultiConnect.ValueBool())
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
	if !m.PassSni.IsNull() && !m.PassSni.IsUnknown() {
		p.PassSni = boolPtr(m.PassSni.ValueBool())
	}
	if !m.PassCipher.IsNull() && !m.PassCipher.IsUnknown() {
		p.PassCipher = boolPtr(m.PassCipher.ValueBool())
	}
	if !m.Verify.IsNull() && !m.Verify.IsUnknown() {
		v := int32(m.Verify.ValueInt64())
		p.Verify = &v
	}
	if !m.ClientCert.IsNull() && !m.ClientCert.IsUnknown() {
		v := int32(m.ClientCert.ValueInt64())
		p.ClientCert = &v
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

func (r *SubVirtualServiceResource) writeState(ctx context.Context, vs *loadmaster.VirtualService, m *SubVirtualServiceResourceModel) diag.Diagnostics {
	m.Id = types.StringValue(strconv.Itoa(int(vs.Index)))
	m.Type = types.StringValue(vs.VSType)
	m.Nickname = types.StringValue(vs.NickName)
	if vs.Enable != nil {
		m.Enabled = types.BoolValue(*vs.Enable)
	} else {
		m.Enabled = types.BoolValue(false)
	}
	m.SSLAcceleration = boolFromPtr(vs.SSLAcceleration)

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
	// m.Persist is intentionally not updated: showvs does not return the persist
	// mode, so we preserve whatever the user last set to avoid perpetual drift.
	m.PersistTimeout = types.StringValue(vs.PersistTimeout)
	m.Idletime = int64FromPtr(vs.Idletime)
	m.ServerInit = int64FromPtr(vs.ServerInit)
	m.ForceL7 = boolFromPtr(vs.ForceL7)
	m.ForceL4 = boolFromPtr(vs.ForceL4)
	m.Transparent = boolFromPtr(vs.Transparent)
	m.UseForSnat = boolFromPtr(vs.UseforSnat)
	m.MultiConnect = boolFromPtr(vs.MultiConnect)
	m.Cache = boolFromPtr(vs.Cache)
	m.Compress = boolFromPtr(vs.Compress)
	m.AllowHTTP2 = boolFromPtr(vs.AllowHTTP2)
	m.SSLReverse = boolFromPtr(vs.SSLReverse)
	m.SSLReencrypt = boolFromPtr(vs.SSLReencrypt)
	m.PassSni = boolFromPtr(vs.PassSni)
	m.PassCipher = boolFromPtr(vs.PassCipher)
	m.Verify = int64FromPtr(vs.Verify)
	m.ClientCert = int64FromPtr(vs.ClientCert)
	m.AddVia = types.StringValue(addViaFromAPI(vs.AddVia))
	m.RefreshPersist = boolFromPtr(vs.RefreshPersist)
	m.RsMinimum = int64FromPtr(vs.RsMinimum)
	m.Bandwidth = int64FromPtr(vs.Bandwidth)
	m.ConnsPerSecLimit = int64FromPtr(vs.ConnsPerSecLimit)
	m.RequestsPerSecLimit = int64FromPtr(vs.RequestsPerSecLimit)
	m.MaxConnsLimit = int64FromPtr(vs.MaxConnsLimit)

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

func (r *SubVirtualServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubVirtualServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.CreateSubVS(ctx, data.ParentId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating sub-virtual service", err.Error())
		return
	}

	params, d := r.paramsFromModel(ctx, data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, mErr := r.client.ModifyVirtualService(ctx, strconv.Itoa(int(vs.Index)), params)
	if mErr != nil {
		resp.Diagnostics.AddError("Error applying sub-virtual service settings post-create", mErr.Error())
		return
	}

	resp.Diagnostics.Append(r.writeState(ctx, updated, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubVirtualServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubVirtualServiceResourceModel
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
		resp.Diagnostics.AddError("Error reading sub-virtual service", err.Error())
		return
	}

	resp.Diagnostics.Append(r.writeState(ctx, vs, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubVirtualServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubVirtualServiceResourceModel
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
		resp.Diagnostics.AddError("Error updating sub-virtual service", err.Error())
		return
	}

	resp.Diagnostics.Append(r.writeState(ctx, vs, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubVirtualServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubVirtualServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVirtualService(ctx, data.Id.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting sub-virtual service", err.Error())
	}
}

// ImportState accepts "<parent_id>/<subvs_id>".
func (r *SubVirtualServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(`expected "<parent_id>/<subvs_id>", got %q`, req.ID))
		return
	}

	vs, err := r.client.ShowVirtualService(ctx, parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Error importing sub-virtual service", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_id"), parts[0])...)
	data := SubVirtualServiceResourceModel{ParentId: types.StringValue(parts[0])}
	resp.Diagnostics.Append(r.writeState(ctx, vs, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

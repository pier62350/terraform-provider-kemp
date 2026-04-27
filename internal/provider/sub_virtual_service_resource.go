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
	PersistTimeout      types.String `tfsdk:"persist_timeout"`
	Idletime            types.Int64  `tfsdk:"idletime"`
	ForceL7             types.Bool   `tfsdk:"force_l7"`
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
	WafInterceptMode    types.String `tfsdk:"waf_intercept_mode"`
	WafBlockingParanoia types.Int64  `tfsdk:"waf_blocking_paranoia"`
	WafAlertThreshold   types.Int64  `tfsdk:"waf_alert_threshold"`
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
			"nickname": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Friendly name shown in the WUI."},
			"enabled":  schema.BoolAttribute{Optional: true, Computed: true},
			"type":     schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "VS type — `gen`, `http`, `http2`, `ts`, `tls`, `log`."},
			"ssl_acceleration": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable SSL/TLS termination on this SubVS.",
			},
			"cert_files": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Names of certificates attached to this SubVS (multiple entries enable SNI).",
			},
			"schedule":                schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Load-balancing algorithm: `rr`, `wlc`, `lc`, `pi`, `ph`, etc."},
			"persist_timeout":         schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Persistence timeout in seconds. `0` disables persistence."},
			"idletime":                schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Idle connection timeout in seconds."},
			"force_l7":                schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Force Layer-7 processing."},
			"bandwidth":               schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Bandwidth limit in Mbps. `0` means unlimited."},
			"conns_per_sec_limit":     schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum new connections per second. `0` means unlimited."},
			"requests_per_sec_limit":  schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum HTTP requests per second. `0` means unlimited."},
			"max_conns_limit":         schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum concurrent connections. `0` means unlimited."},
			"check_type":              schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Health check type: `tcp`, `http`, `https`, `icmp`, `none`, etc."},
			"check_port":              schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Port used for health checks. `0` means use the VS port."},
			"chk_interval":            schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Interval between health checks in seconds."},
			"chk_timeout":             schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Health check timeout in seconds."},
			"chk_retry_count":         schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Consecutive failures before marking a real server down."},
			"need_host_name":          schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Send the VS hostname in the HTTP `Host` header during health checks."},
			"check_use_http11":        schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Use HTTP/1.1 for HTTP-based health checks."},
			"check_use_get":           schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "HTTP method for health checks: `head` (default) or `get`."},
			"match_len":               schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Bytes of health check response body to inspect for match pattern."},
			"enhanced_health_checks":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Enable enhanced health checks."},
			"esp_enabled":               schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Enable Edge Security Pack on this SubVS."},
			"esp_allowed_hosts":         schema.StringAttribute{Optional: true, Computed: true},
			"esp_allowed_directories":   schema.StringAttribute{Optional: true, Computed: true},
			"esp_input_auth_mode":       schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Client-side auth mode: `none`, `basic`, `form`."},
			"esp_output_auth_mode":      schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Server-side auth mode: `none`, `basic`, `form`, `kcd`."},
			"esp_include_nested_groups": schema.BoolAttribute{Optional: true, Computed: true},
			"esp_display_pub_priv":      schema.BoolAttribute{Optional: true, Computed: true},
			"esp_logs":                  schema.BoolAttribute{Optional: true, Computed: true},
			"waf_intercept_mode":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "WAF intercept mode: `disabled`, `legacy`, or `owasp`."},
			"waf_blocking_paranoia":     schema.Int64Attribute{Optional: true, Computed: true},
			"waf_alert_threshold":       schema.Int64Attribute{Optional: true, Computed: true},
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
	m.PersistTimeout = types.StringValue(vs.PersistTimeout)
	m.Idletime = int64FromPtr(vs.Idletime)
	m.ForceL7 = boolFromPtr(vs.ForceL7)
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

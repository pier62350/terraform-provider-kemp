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

	EspEnabled             types.Bool   `tfsdk:"esp_enabled"`
	EspAllowedHosts        types.String `tfsdk:"esp_allowed_hosts"`
	EspAllowedDirectories  types.String `tfsdk:"esp_allowed_directories"`
	EspInputAuthMode       types.String `tfsdk:"esp_input_auth_mode"`
	EspOutputAuthMode      types.String `tfsdk:"esp_output_auth_mode"`
	EspIncludeNestedGroups types.Bool   `tfsdk:"esp_include_nested_groups"`
	EspDisplayPubPriv      types.Bool   `tfsdk:"esp_display_pub_priv"`
	EspLogs                types.Bool   `tfsdk:"esp_logs"`

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
			"esp_enabled":               schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Enable Edge Security Pack on this SubVS."},
			"esp_allowed_hosts":         schema.StringAttribute{Optional: true, Computed: true},
			"esp_allowed_directories":   schema.StringAttribute{Optional: true, Computed: true},
			"esp_input_auth_mode":       schema.StringAttribute{Optional: true, Computed: true},
			"esp_output_auth_mode":      schema.StringAttribute{Optional: true, Computed: true},
			"esp_include_nested_groups": schema.BoolAttribute{Optional: true, Computed: true},
			"esp_display_pub_priv":      schema.BoolAttribute{Optional: true, Computed: true},
			"esp_logs":                  schema.BoolAttribute{Optional: true, Computed: true},
			"waf_intercept_mode":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "WAF intercept mode: `0` off / `1` Legacy / `2` OWASP."},
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
		p.InputAuthMode = m.EspInputAuthMode.ValueString()
	}
	if !m.EspOutputAuthMode.IsNull() && !m.EspOutputAuthMode.IsUnknown() {
		p.OutputAuthMode = m.EspOutputAuthMode.ValueString()
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
		p.InterceptMode = m.WafInterceptMode.ValueString()
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

	m.EspEnabled = boolFromPtr(vs.EspEnabled)
	m.EspAllowedHosts = types.StringValue(vs.AllowedHosts)
	m.EspAllowedDirectories = types.StringValue(vs.AllowedDirectories)
	m.EspInputAuthMode = types.StringValue(vs.InputAuthMode)
	m.EspOutputAuthMode = types.StringValue(vs.OutputAuthMode)
	m.EspIncludeNestedGroups = boolFromPtr(vs.IncludeNestedGroups)
	m.EspDisplayPubPriv = boolFromPtr(vs.DisplayPubPriv)
	m.EspLogs = boolFromPtr(vs.EspLogs)

	m.WafInterceptMode = types.StringValue(vs.InterceptMode)
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

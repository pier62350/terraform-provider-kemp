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

	// ESP
	EspEnabled             types.Bool   `tfsdk:"esp_enabled"`
	EspAllowedHosts        types.String `tfsdk:"esp_allowed_hosts"`
	EspAllowedDirectories  types.String `tfsdk:"esp_allowed_directories"`
	EspInputAuthMode       types.String `tfsdk:"esp_input_auth_mode"`
	EspOutputAuthMode      types.String `tfsdk:"esp_output_auth_mode"`
	EspIncludeNestedGroups types.Bool   `tfsdk:"esp_include_nested_groups"`
	EspDisplayPubPriv      types.Bool   `tfsdk:"esp_display_pub_priv"`
	EspLogs                types.Bool   `tfsdk:"esp_logs"`
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
				MarkdownDescription: "Client-side authentication mode (e.g. `0` none, `1` basic auth, `2` form-based). Refer to LoadMaster docs for the full enum.",
				Optional:            true,
				Computed:            true,
			},
			"esp_output_auth_mode": schema.StringAttribute{
				MarkdownDescription: "Server-side authentication mode for the upstream (e.g. `0` none, `1` basic, `2` form, `4` KCD).",
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

	m.EspEnabled = boolFromPtr(vs.EspEnabled)
	m.EspAllowedHosts = types.StringValue(vs.AllowedHosts)
	m.EspAllowedDirectories = types.StringValue(vs.AllowedDirectories)
	m.EspInputAuthMode = types.StringValue(vs.InputAuthMode)
	m.EspOutputAuthMode = types.StringValue(vs.OutputAuthMode)
	m.EspIncludeNestedGroups = boolFromPtr(vs.IncludeNestedGroups)
	m.EspDisplayPubPriv = boolFromPtr(vs.DisplayPubPriv)
	m.EspLogs = boolFromPtr(vs.EspLogs)

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

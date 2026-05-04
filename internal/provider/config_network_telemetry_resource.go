// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigNetworkTelemetryResource{}
	_ resource.ResourceWithImportState = &ConfigNetworkTelemetryResource{}
)

func NewConfigNetworkTelemetryResource() resource.Resource {
	return &ConfigNetworkTelemetryResource{}
}

type ConfigNetworkTelemetryResource struct {
	client *loadmaster.Client
}

type ConfigNetworkTelemetryModel struct {
	InterfaceID types.String `tfsdk:"interface_id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

func (r *ConfigNetworkTelemetryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_network_telemetry"
}

func (r *ConfigNetworkTelemetryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the telemetry enabled state for a network interface on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"interface_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Network interface ID (e.g. `\"1\"`). Acts as the resource ID. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Whether telemetry is enabled for this interface.",
			},
		},
	}
}

func (r *ConfigNetworkTelemetryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigNetworkTelemetryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigNetworkTelemetryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetTelemetry(ctx, data.InterfaceID.ValueString(), data.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Error configuring network telemetry", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigNetworkTelemetryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigNetworkTelemetryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled, err := r.client.GetTelemetry(ctx, data.InterfaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading network telemetry", err.Error())
		return
	}
	data.Enabled = types.BoolValue(enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigNetworkTelemetryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigNetworkTelemetryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetTelemetry(ctx, data.InterfaceID.ValueString(), data.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Error updating network telemetry", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigNetworkTelemetryResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Telemetry state is per-interface config — removing from state is sufficient.
}

func (r *ConfigNetworkTelemetryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	enabled, err := r.client.GetTelemetry(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing network telemetry", err.Error())
		return
	}
	data := ConfigNetworkTelemetryModel{
		InterfaceID: types.StringValue(req.ID),
		Enabled:     types.BoolValue(enabled),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ resource.Resource                = &ConfigGEOResource{}
	_ resource.ResourceWithImportState = &ConfigGEOResource{}
)

func NewConfigGEOResource() resource.Resource {
	return &ConfigGEOResource{}
}

type ConfigGEOResource struct {
	client *loadmaster.Client
}

type ConfigGEOModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (r *ConfigGEOResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_geo"
}

func (r *ConfigGEOResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the GEO (Global Server Load Balancing) enabled state on a Kemp LoadMaster.

This is a singleton resource. Import with: ` + "`terraform import kemp_config_geo.main loadmaster`",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `loadmaster`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Whether GEO is enabled on the LoadMaster.",
			},
		},
	}
}

func (r *ConfigGEOResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigGEOResource) applyEnabled(ctx context.Context, enabled bool) error {
	if enabled {
		return r.client.EnableGEO(ctx)
	}
	return r.client.DisableGEO(ctx)
}

func (r *ConfigGEOResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigGEOModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyEnabled(ctx, data.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Error configuring GEO", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigGEOResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigGEOModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled, err := r.client.IsGEOEnabled(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading GEO state", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	data.Enabled = types.BoolValue(enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigGEOResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigGEOModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyEnabled(ctx, data.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Error updating GEO state", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigGEOResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton — removing from state is sufficient; LoadMaster config is unchanged.
}

func (r *ConfigGEOResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	enabled, err := r.client.IsGEOEnabled(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing GEO state", err.Error())
		return
	}
	data := ConfigGEOModel{
		ID:      types.StringValue("loadmaster"),
		Enabled: types.BoolValue(enabled),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

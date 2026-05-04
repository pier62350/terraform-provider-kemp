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
	_ resource.Resource                = &ConfigIntrusionDetectionResource{}
	_ resource.ResourceWithImportState = &ConfigIntrusionDetectionResource{}
)

func NewConfigIntrusionDetectionResource() resource.Resource {
	return &ConfigIntrusionDetectionResource{}
}

type ConfigIntrusionDetectionResource struct {
	client *loadmaster.Client
}

type ConfigIntrusionDetectionResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Level types.Int64  `tfsdk:"level"`
}

func (r *ConfigIntrusionDetectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_intrusion_detection"
}

func (r *ConfigIntrusionDetectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the IPS/IDS intrusion detection (paranoia) level on a Kemp LoadMaster.

This is a singleton resource. Import with: ` + "`terraform import kemp_config_intrusion_detection.main loadmaster`",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `loadmaster`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"level": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** IPS/IDS paranoia detection level. Range: 0–4.",
			},
		},
	}
}

func (r *ConfigIntrusionDetectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigIntrusionDetectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigIntrusionDetectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetIntrusionDetectionLevel(ctx, data.Level.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error configuring intrusion detection level", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigIntrusionDetectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigIntrusionDetectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	level, err := r.client.GetIntrusionDetectionLevel(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading intrusion detection level", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	data.Level = types.Int64Value(level)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigIntrusionDetectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigIntrusionDetectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetIntrusionDetectionLevel(ctx, data.Level.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error updating intrusion detection level", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigIntrusionDetectionResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton — removing from state is sufficient; LoadMaster config is unchanged.
}

func (r *ConfigIntrusionDetectionResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	level, err := r.client.GetIntrusionDetectionLevel(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing intrusion detection level", err.Error())
		return
	}
	data := ConfigIntrusionDetectionResourceModel{
		ID:    types.StringValue("loadmaster"),
		Level: types.Int64Value(level),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

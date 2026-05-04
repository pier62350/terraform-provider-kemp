// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigRPSLimitResource{}
	_ resource.ResourceWithImportState = &ConfigRPSLimitResource{}
)

func NewConfigRPSLimitResource() resource.Resource {
	return &ConfigRPSLimitResource{}
}

type ConfigRPSLimitResource struct {
	client *loadmaster.Client
}

type ConfigRPSLimitModel struct {
	Address types.String `tfsdk:"address"`
	Limit   types.Int64  `tfsdk:"limit"`
}

func (r *ConfigRPSLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_rps_limit"
}

func (r *ConfigRPSLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a global client requests-per-second (RPS) limit per IP address on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Client IP address or CIDR to rate-limit. Acts as the resource ID. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"limit": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** Maximum HTTP requests per second allowed from this address.",
			},
		},
	}
}

func (r *ConfigRPSLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigRPSLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigRPSLimitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddRPSLimit(ctx, data.Address.ValueString(), strconv.FormatInt(data.Limit.ValueInt64(), 10)); err != nil {
		resp.Diagnostics.AddError("Error creating RPS limit", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigRPSLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigRPSLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := r.client.ListRPSLimits(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading RPS limits", err.Error())
		return
	}

	addr := data.Address.ValueString()
	for _, entry := range entries {
		if entry.Addr == addr {
			data.Limit = types.Int64Value(entry.Limit)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *ConfigRPSLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan ConfigRPSLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRPSLimit(ctx, state.Address.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error updating RPS limit", err.Error())
		return
	}
	if err := r.client.AddRPSLimit(ctx, plan.Address.ValueString(), strconv.FormatInt(plan.Limit.ValueInt64(), 10)); err != nil {
		resp.Diagnostics.AddError("Error updating RPS limit", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigRPSLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigRPSLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRPSLimit(ctx, data.Address.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting RPS limit", err.Error())
	}
}

func (r *ConfigRPSLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	entries, err := r.client.ListRPSLimits(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing RPS limit", err.Error())
		return
	}

	for _, entry := range entries {
		if entry.Addr == req.ID {
			data := ConfigRPSLimitModel{
				Address: types.StringValue(entry.Addr),
				Limit:   types.Int64Value(entry.Limit),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.Diagnostics.AddError("RPS limit not found", fmt.Sprintf("No RPS limit for address %q found.", req.ID))
}

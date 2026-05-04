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
	_ resource.Resource                = &ConfigBandwidthLimitResource{}
	_ resource.ResourceWithImportState = &ConfigBandwidthLimitResource{}
)

func NewConfigBandwidthLimitResource() resource.Resource {
	return &ConfigBandwidthLimitResource{}
}

type ConfigBandwidthLimitResource struct {
	client *loadmaster.Client
}

type ConfigBandwidthLimitModel struct {
	Address types.String `tfsdk:"address"`
	Limit   types.Int64  `tfsdk:"limit"`
}

func (r *ConfigBandwidthLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_bandwidth_limit"
}

func (r *ConfigBandwidthLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a global client bandwidth limit per IP address on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Client IP address or CIDR to rate-limit. Acts as the resource ID. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"limit": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** Bandwidth limit value (in Kbps).",
			},
		},
	}
}

func (r *ConfigBandwidthLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigBandwidthLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigBandwidthLimitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddBandwidthLimit(ctx, data.Address.ValueString(), strconv.FormatInt(data.Limit.ValueInt64(), 10)); err != nil {
		resp.Diagnostics.AddError("Error creating bandwidth limit", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigBandwidthLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigBandwidthLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := r.client.ListBandwidthLimits(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading bandwidth limits", err.Error())
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

func (r *ConfigBandwidthLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan ConfigBandwidthLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteBandwidthLimit(ctx, state.Address.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error updating bandwidth limit", err.Error())
		return
	}
	if err := r.client.AddBandwidthLimit(ctx, plan.Address.ValueString(), strconv.FormatInt(plan.Limit.ValueInt64(), 10)); err != nil {
		resp.Diagnostics.AddError("Error updating bandwidth limit", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigBandwidthLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigBandwidthLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteBandwidthLimit(ctx, data.Address.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting bandwidth limit", err.Error())
	}
}

func (r *ConfigBandwidthLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	entries, err := r.client.ListBandwidthLimits(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing bandwidth limit", err.Error())
		return
	}

	for _, entry := range entries {
		if entry.Addr == req.ID {
			data := ConfigBandwidthLimitModel{
				Address: types.StringValue(entry.Addr),
				Limit:   types.Int64Value(entry.Limit),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.Diagnostics.AddError("Bandwidth limit not found", fmt.Sprintf("No bandwidth limit for address %q found.", req.ID))
}

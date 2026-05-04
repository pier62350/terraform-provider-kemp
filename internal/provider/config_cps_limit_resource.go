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
	_ resource.Resource                = &ConfigCPSLimitResource{}
	_ resource.ResourceWithImportState = &ConfigCPSLimitResource{}
)

func NewConfigCPSLimitResource() resource.Resource {
	return &ConfigCPSLimitResource{}
}

type ConfigCPSLimitResource struct {
	client *loadmaster.Client
}

type ConfigCPSLimitModel struct {
	Address types.String `tfsdk:"address"`
	Limit   types.Int64  `tfsdk:"limit"`
}

func (r *ConfigCPSLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_cps_limit"
}

func (r *ConfigCPSLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a global client connections-per-second (CPS) limit per IP address on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Client IP address or CIDR to rate-limit. Acts as the resource ID. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"limit": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** Maximum new connections per second allowed from this address.",
			},
		},
	}
}

func (r *ConfigCPSLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigCPSLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigCPSLimitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddCPSLimit(ctx, data.Address.ValueString(), strconv.FormatInt(data.Limit.ValueInt64(), 10)); err != nil {
		resp.Diagnostics.AddError("Error creating CPS limit", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigCPSLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigCPSLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := r.client.ListCPSLimits(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CPS limits", err.Error())
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

func (r *ConfigCPSLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan ConfigCPSLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCPSLimit(ctx, state.Address.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error updating CPS limit", err.Error())
		return
	}
	if err := r.client.AddCPSLimit(ctx, plan.Address.ValueString(), strconv.FormatInt(plan.Limit.ValueInt64(), 10)); err != nil {
		resp.Diagnostics.AddError("Error updating CPS limit", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigCPSLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigCPSLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCPSLimit(ctx, data.Address.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting CPS limit", err.Error())
	}
}

func (r *ConfigCPSLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	entries, err := r.client.ListCPSLimits(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing CPS limit", err.Error())
		return
	}

	for _, entry := range entries {
		if entry.Addr == req.ID {
			data := ConfigCPSLimitModel{
				Address: types.StringValue(entry.Addr),
				Limit:   types.Int64Value(entry.Limit),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.Diagnostics.AddError("CPS limit not found", fmt.Sprintf("No CPS limit for address %q found.", req.ID))
}

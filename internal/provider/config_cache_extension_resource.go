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
	_ resource.Resource                = &ConfigCacheExtensionResource{}
	_ resource.ResourceWithImportState = &ConfigCacheExtensionResource{}
)

func NewConfigCacheExtensionResource() resource.Resource {
	return &ConfigCacheExtensionResource{}
}

type ConfigCacheExtensionResource struct {
	client *loadmaster.Client
}

type ConfigCacheExtensionModel struct {
	Extension types.String `tfsdk:"extension"`
}

func (r *ConfigCacheExtensionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_cache_extension"
}

func (r *ConfigCacheExtensionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a no-cache file extension on a Kemp LoadMaster. Files matching this extension will not be cached.",
		Attributes: map[string]schema.Attribute{
			"extension": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** File extension to exclude from caching (e.g. `.mp4`). Acts as the resource ID. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ConfigCacheExtensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigCacheExtensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigCacheExtensionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddNoCacheExtension(ctx, data.Extension.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating no-cache extension", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is a no-op: the LoadMaster API provides no list command for no-cache extensions.
// Drift will not be detected, but create and delete work correctly.
func (r *ConfigCacheExtensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigCacheExtensionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigCacheExtensionResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All attributes are ForceNew — no in-place update path.
}

func (r *ConfigCacheExtensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigCacheExtensionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNoCacheExtension(ctx, data.Extension.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting no-cache extension", err.Error())
	}
}

func (r *ConfigCacheExtensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Since Read is a no-op, simply restore extension from the import ID.
	data := ConfigCacheExtensionModel{
		Extension: types.StringValue(req.ID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

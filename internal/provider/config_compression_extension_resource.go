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
	_ resource.Resource                = &ConfigCompressionExtensionResource{}
	_ resource.ResourceWithImportState = &ConfigCompressionExtensionResource{}
)

func NewConfigCompressionExtensionResource() resource.Resource {
	return &ConfigCompressionExtensionResource{}
}

type ConfigCompressionExtensionResource struct {
	client *loadmaster.Client
}

type ConfigCompressionExtensionModel struct {
	Extension types.String `tfsdk:"extension"`
}

func (r *ConfigCompressionExtensionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_compression_extension"
}

func (r *ConfigCompressionExtensionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a no-compress file extension on a Kemp LoadMaster. Files matching this extension will not be compressed.",
		Attributes: map[string]schema.Attribute{
			"extension": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** File extension to exclude from compression (e.g. `.mp4`). Acts as the resource ID. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ConfigCompressionExtensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigCompressionExtensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigCompressionExtensionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddNoCompressExtension(ctx, data.Extension.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating no-compress extension", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is a no-op: the LoadMaster API provides no list command for no-compress extensions.
// Drift will not be detected, but create and delete work correctly.
func (r *ConfigCompressionExtensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigCompressionExtensionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigCompressionExtensionResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All attributes are ForceNew — no in-place update path.
}

func (r *ConfigCompressionExtensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigCompressionExtensionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNoCompressExtension(ctx, data.Extension.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting no-compress extension", err.Error())
	}
}

func (r *ConfigCompressionExtensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Since Read is a no-op, simply restore extension from the import ID.
	data := ConfigCompressionExtensionModel{
		Extension: types.StringValue(req.ID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

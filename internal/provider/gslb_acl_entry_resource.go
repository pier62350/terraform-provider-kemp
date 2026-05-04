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
	_ resource.Resource                = &GSLBACLEntryResource{}
	_ resource.ResourceWithImportState = &GSLBACLEntryResource{}
)

func NewGSLBACLEntryResource() resource.Resource { return &GSLBACLEntryResource{} }

type GSLBACLEntryResource struct {
	client *loadmaster.Client
}

type GSLBACLEntryResourceModel struct {
	Address types.String `tfsdk:"address"`
}

func (r *GSLBACLEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gslb_acl_entry"
}

func (r *GSLBACLEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an IP address or CIDR in the GEO feature user-defined custom allow-list on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** IP address or CIDR to add to the GEO custom allow-list. Acts as the resource identity. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *GSLBACLEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GSLBACLEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GSLBACLEntryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddGSLBACLEntry(ctx, data.Address.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating GSLB ACL entry", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBACLEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GSLBACLEntryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := r.client.ListGSLBACLEntries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading GSLB ACL entries", err.Error())
		return
	}

	for _, e := range entries {
		if e == data.Address.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	// Not found in the list — remove from state.
	resp.State.RemoveResource(ctx)
}

func (r *GSLBACLEntryResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// address is ForceNew — no in-place update possible.
}

func (r *GSLBACLEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GSLBACLEntryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGSLBACLEntry(ctx, data.Address.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting GSLB ACL entry", err.Error())
	}
}

func (r *GSLBACLEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	entries, err := r.client.ListGSLBACLEntries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing GSLB ACL entry", err.Error())
		return
	}

	for _, e := range entries {
		if e == req.ID {
			data := GSLBACLEntryResourceModel{Address: types.StringValue(req.ID)}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.Diagnostics.AddError("GSLB ACL entry not found", fmt.Sprintf("No GSLB ACL entry with address %q found.", req.ID))
}

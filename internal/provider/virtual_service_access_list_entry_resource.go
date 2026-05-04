// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &VirtualServiceAccessListEntryResource{}
	_ resource.ResourceWithImportState = &VirtualServiceAccessListEntryResource{}
)

func NewVirtualServiceAccessListEntryResource() resource.Resource {
	return &VirtualServiceAccessListEntryResource{}
}

type VirtualServiceAccessListEntryResource struct {
	client *loadmaster.Client
}

type VirtualServiceAccessListEntryModel struct {
	VirtualServiceId types.String `tfsdk:"virtual_service_id"`
	ListType         types.String `tfsdk:"list_type"`
	Address          types.String `tfsdk:"address"`
}

func (r *VirtualServiceAccessListEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_service_access_list_entry"
}

func (r *VirtualServiceAccessListEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an entry in a per-virtual-service allow or block access list on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Numeric index of the virtual service. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"list_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of access list: `allow` or `block`. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IP address to add to the list (e.g. `10.0.0.5`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *VirtualServiceAccessListEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtualServiceAccessListEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtualServiceAccessListEntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddVSACLEntry(ctx,
		data.VirtualServiceId.ValueString(),
		data.ListType.ValueString(),
		data.Address.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error creating virtual service access list entry", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceAccessListEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtualServiceAccessListEntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entries, err := r.client.ListVSACLEntries(ctx, data.VirtualServiceId.ValueString(), data.ListType.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading virtual service access list entries", err.Error())
		return
	}
	addr := data.Address.ValueString()
	for _, e := range entries {
		if e == addr {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *VirtualServiceAccessListEntryResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are ForceNew; framework should never call Update.
	resp.Diagnostics.AddError("Update not supported", "kemp_virtual_service_access_list_entry has no in-place updatable attributes; changes trigger replacement.")
}

func (r *VirtualServiceAccessListEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtualServiceAccessListEntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVSACLEntry(ctx,
		data.VirtualServiceId.ValueString(),
		data.ListType.ValueString(),
		data.Address.ValueString(),
	); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting virtual service access list entry", err.Error())
	}
}

// ImportState accepts "<virtual_service_id>/<list_type>/<address>".
func (r *VirtualServiceAccessListEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(`expected "<virtual_service_id>/<list_type>/<address>", got %q`, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_service_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("list_type"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("address"), parts[2])...)
}

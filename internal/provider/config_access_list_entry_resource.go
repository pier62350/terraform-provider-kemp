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
	_ resource.Resource                = &ConfigAccessListEntryResource{}
	_ resource.ResourceWithImportState = &ConfigAccessListEntryResource{}
)

func NewConfigAccessListEntryResource() resource.Resource { return &ConfigAccessListEntryResource{} }

type ConfigAccessListEntryResource struct {
	client *loadmaster.Client
}

type ConfigAccessListEntryModel struct {
	ListType types.String `tfsdk:"list_type"`
	Address  types.String `tfsdk:"address"`
}

func (r *ConfigAccessListEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_access_list_entry"
}

func (r *ConfigAccessListEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an entry in the appliance-wide (global) allow or block access list on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"list_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of access list: `allow` or `block`. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IP address or CIDR to add to the list (e.g. `10.0.0.1` or `192.168.1.0/24`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ConfigAccessListEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigAccessListEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigAccessListEntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddGlobalACLEntry(ctx, data.ListType.ValueString(), data.Address.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating global access list entry", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigAccessListEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigAccessListEntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entries, err := r.client.ListGlobalACLEntries(ctx, data.ListType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading global access list entries", err.Error())
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

func (r *ConfigAccessListEntryResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are ForceNew; framework should never call Update.
	resp.Diagnostics.AddError("Update not supported", "kemp_config_access_list_entry has no in-place updatable attributes; changes trigger replacement.")
}

func (r *ConfigAccessListEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigAccessListEntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGlobalACLEntry(ctx, data.ListType.ValueString(), data.Address.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting global access list entry", err.Error())
	}
}

// ImportState accepts "<list_type>/<address>" (e.g. "allow/10.0.0.1").
func (r *ConfigAccessListEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(`expected "<list_type>/<address>", got %q`, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("list_type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("address"), parts[1])...)
}

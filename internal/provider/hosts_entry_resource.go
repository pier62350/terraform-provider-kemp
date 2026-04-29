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
	_ resource.Resource                = &HostsEntryResource{}
	_ resource.ResourceWithImportState = &HostsEntryResource{}
)

func NewHostsEntryResource() resource.Resource { return &HostsEntryResource{} }

type HostsEntryResource struct {
	client *loadmaster.Client
}

type HostsEntryResourceModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	FQDN      types.String `tfsdk:"fqdn"`
}

func (r *HostsEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosts_entry"
}

func (r *HostsEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a hosts file entry on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"ip_address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** IP address for the hosts entry. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"fqdn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Fully qualified domain name for the hosts entry. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *HostsEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *HostsEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HostsEntryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddHostsEntry(ctx, data.IPAddress.ValueString(), data.FQDN.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating hosts entry", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostsEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HostsEntryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := r.client.ListHostsEntries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading hosts entries", err.Error())
		return
	}

	ip := data.IPAddress.ValueString()
	for _, entry := range entries {
		if entry.HostIPAddress == ip {
			data.FQDN = types.StringValue(entry.HostFqdn)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *HostsEntryResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All attributes are ForceNew — no in-place update path.
}

func (r *HostsEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HostsEntryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteHostsEntry(ctx, data.IPAddress.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting hosts entry", err.Error())
	}
}

func (r *HostsEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	entries, err := r.client.ListHostsEntries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing hosts entry", err.Error())
		return
	}

	for _, entry := range entries {
		if entry.HostIPAddress == req.ID {
			data := HostsEntryResourceModel{
				IPAddress: types.StringValue(entry.HostIPAddress),
				FQDN:      types.StringValue(entry.HostFqdn),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.Diagnostics.AddError("Hosts entry not found", fmt.Sprintf("No hosts entry with IP %q found.", req.ID))
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &GroupResource{}
	_ resource.ResourceWithImportState = &GroupResource{}
)

func NewGroupResource() resource.Resource { return &GroupResource{} }

type GroupResource struct {
	client *loadmaster.Client
}

type GroupResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Permissions types.List   `tfsdk:"permissions"`
}

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a remote user group on a Kemp LoadMaster. Remote groups are used to map LDAP/Active Directory groups to LoadMaster permissions.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Name of the remote group. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. List of permissions granted to the group. Valid values: `real`, `vs`, `rules`, `backup`, `certs`. An empty list means read-only access.",
				PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func groupPermsToWire(ctx context.Context, perms types.List) string {
	if perms.IsNull() || perms.IsUnknown() {
		return ""
	}
	var items []string
	_ = perms.ElementsAs(ctx, &items, false)
	return strings.Join(items, ",")
}

func groupPermsFromWire(ctx context.Context, wire string) types.List {
	wire = strings.TrimSpace(wire)
	if wire == "" || strings.EqualFold(wire, "ReadOnly") {
		v, _ := types.ListValueFrom(ctx, types.StringType, []string{})
		return v
	}
	items := strings.Split(wire, ",")
	for i, item := range items {
		items[i] = strings.TrimSpace(item)
	}
	v, _ := types.ListValueFrom(ctx, types.StringType, items)
	return v
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddRemoteGroup(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating remote group", err.Error())
		return
	}

	wire := groupPermsToWire(ctx, data.Permissions)
	if wire != "" {
		if err := r.client.SetGroupPerms(ctx, data.Name.ValueString(), wire); err != nil {
			resp.Diagnostics.AddError("Error setting group permissions", err.Error())
			return
		}
	}

	grp, err := r.client.ShowRemoteGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading remote group after create", err.Error())
		return
	}
	data.Permissions = groupPermsFromWire(ctx, grp.Perms)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grp, err := r.client.ShowRemoteGroup(ctx, data.Name.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading remote group", err.Error())
		return
	}
	data.Permissions = groupPermsFromWire(ctx, grp.Perms)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wire := groupPermsToWire(ctx, data.Permissions)
	if err := r.client.SetGroupPerms(ctx, data.Name.ValueString(), wire); err != nil {
		resp.Diagnostics.AddError("Error updating group permissions", err.Error())
		return
	}

	grp, err := r.client.ShowRemoteGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading remote group after update", err.Error())
		return
	}
	data.Permissions = groupPermsFromWire(ctx, grp.Perms)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRemoteGroup(ctx, data.Name.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting remote group", err.Error())
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	grp, err := r.client.ShowRemoteGroup(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing remote group", err.Error())
		return
	}
	data := GroupResourceModel{
		Name:        types.StringValue(grp.Name),
		Permissions: groupPermsFromWire(ctx, grp.Perms),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ resource.Resource                = &LocalUserResource{}
	_ resource.ResourceWithImportState = &LocalUserResource{}
)

func NewLocalUserResource() resource.Resource { return &LocalUserResource{} }

type LocalUserResource struct {
	client *loadmaster.Client
}

type LocalUserResourceModel struct {
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Permissions types.List   `tfsdk:"permissions"`
}

func (r *LocalUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_local_user"
}

func (r *LocalUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a local user account on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Username for the local account. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "**Required.** Password for the local account. Write-only — not read back from the API. After `terraform import`, this attribute will be null in state and will show a diff until set in configuration.",
			},
			"permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. List of permissions granted to the user. Valid values: `real`, `vs`, `rules`, `backup`, `certs`. An empty list means read-only access.",
				PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *LocalUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func permsToWire(ctx context.Context, perms types.List) string {
	if perms.IsNull() || perms.IsUnknown() {
		return ""
	}
	var items []string
	_ = perms.ElementsAs(ctx, &items, false)
	return strings.Join(items, ",")
}

func permsFromWire(ctx context.Context, wire string) types.List {
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

func (r *LocalUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LocalUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddLocalUser(ctx, data.Username.ValueString(), data.Password.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating local user", err.Error())
		return
	}

	wire := permsToWire(ctx, data.Permissions)
	if wire != "" {
		if err := r.client.SetUserPerms(ctx, data.Username.ValueString(), wire); err != nil {
			resp.Diagnostics.AddError("Error setting user permissions", err.Error())
			return
		}
	}

	user, err := r.client.ShowLocalUser(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading local user after create", err.Error())
		return
	}
	data.Permissions = permsFromWire(ctx, user.Perms)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocalUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LocalUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.ShowLocalUser(ctx, data.Username.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading local user", err.Error())
		return
	}
	data.Permissions = permsFromWire(ctx, user.Perms)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocalUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LocalUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wire := permsToWire(ctx, data.Permissions)
	if err := r.client.SetUserPerms(ctx, data.Username.ValueString(), wire); err != nil {
		resp.Diagnostics.AddError("Error updating user permissions", err.Error())
		return
	}

	user, err := r.client.ShowLocalUser(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading local user after update", err.Error())
		return
	}
	data.Permissions = permsFromWire(ctx, user.Perms)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocalUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LocalUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteLocalUser(ctx, data.Username.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting local user", err.Error())
	}
}

func (r *LocalUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	user, err := r.client.ShowLocalUser(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing local user", err.Error())
		return
	}
	data := LocalUserResourceModel{
		Username:    types.StringValue(user.Name),
		Password:    types.StringNull(),
		Permissions: permsFromWire(ctx, user.Perms),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ resource.Resource                = &ConfigInterfaceBondResource{}
	_ resource.ResourceWithImportState = &ConfigInterfaceBondResource{}
)

func NewConfigInterfaceBondResource() resource.Resource { return &ConfigInterfaceBondResource{} }

type ConfigInterfaceBondResource struct {
	client *loadmaster.Client
}

type ConfigInterfaceBondResourceModel struct {
	InterfaceID types.String `tfsdk:"interface_id"`
	Members     types.List   `tfsdk:"members"`
}

func (r *ConfigInterfaceBondResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_interface_bond"
}

func (r *ConfigInterfaceBondResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a bonded (LAG) network interface on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"interface_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Interface index of the bond primary interface (e.g. `\"1\"`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"members": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Optional. Interface IDs to include as bond members (e.g. `[\"2\", \"3\"]`).",
			},
		},
	}
}

func (r *ConfigInterfaceBondResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func membersFromModel(ctx context.Context, data ConfigInterfaceBondResourceModel) []string {
	if data.Members.IsNull() || data.Members.IsUnknown() {
		return nil
	}
	var raw []string
	data.Members.ElementsAs(ctx, &raw, false)
	return raw
}

func (r *ConfigInterfaceBondResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigInterfaceBondResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ifaceID := data.InterfaceID.ValueString()
	if err := r.client.CreateBond(ctx, ifaceID); err != nil {
		resp.Diagnostics.AddError("Error creating bond", err.Error())
		return
	}

	for _, member := range membersFromModel(ctx, data) {
		if err := r.client.AddBondMember(ctx, ifaceID, member); err != nil {
			resp.Diagnostics.AddError("Error adding bond member", fmt.Sprintf("Failed to add interface %q to bond %q: %s.", member, ifaceID, err))
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigInterfaceBondResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigInterfaceBondResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface, err := r.client.ShowInterface(ctx, data.InterfaceID.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading interface", err.Error())
		return
	}

	// Verify it's still a bond by checking InterfaceType.
	if iface.InterfaceType != "bond" && iface.InterfaceType != "Bond" && iface.InterfaceType != "bonding" {
		// No longer a bond — remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	// Members cannot be reliably read — preserve state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigInterfaceBondResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ConfigInterfaceBondResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ifaceID := plan.InterfaceID.ValueString()
	planMembers := membersFromModel(ctx, plan)
	stateMembers := membersFromModel(ctx, state)

	// Build sets for comparison.
	stateSet := make(map[string]struct{}, len(stateMembers))
	for _, m := range stateMembers {
		stateSet[m] = struct{}{}
	}
	planSet := make(map[string]struct{}, len(planMembers))
	for _, m := range planMembers {
		planSet[m] = struct{}{}
	}

	// Remove members no longer in plan.
	for _, m := range stateMembers {
		if _, ok := planSet[m]; !ok {
			if err := r.client.DeleteBondMember(ctx, ifaceID, m); err != nil && !loadmaster.IsNotFound(err) {
				resp.Diagnostics.AddError("Error removing bond member", fmt.Sprintf("Failed to remove interface %q from bond %q: %s.", m, ifaceID, err))
				return
			}
		}
	}

	// Add new members.
	for _, m := range planMembers {
		if _, ok := stateSet[m]; !ok {
			if err := r.client.AddBondMember(ctx, ifaceID, m); err != nil {
				resp.Diagnostics.AddError("Error adding bond member", fmt.Sprintf("Failed to add interface %q to bond %q: %s.", m, ifaceID, err))
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigInterfaceBondResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigInterfaceBondResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ifaceID := data.InterfaceID.ValueString()
	for _, member := range membersFromModel(ctx, data) {
		if err := r.client.DeleteBondMember(ctx, ifaceID, member); err != nil && !loadmaster.IsNotFound(err) {
			resp.Diagnostics.AddError("Error removing bond member", fmt.Sprintf("Failed to remove interface %q from bond %q: %s.", member, ifaceID, err))
			return
		}
	}

	if err := r.client.UnbondInterface(ctx, ifaceID); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error unbonding interface", err.Error())
	}
}

func (r *ConfigInterfaceBondResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	interfaceID := req.ID
	iface, err := r.client.ShowInterface(ctx, interfaceID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing bond", err.Error())
		return
	}
	if iface == nil {
		resp.Diagnostics.AddError("Interface not found", fmt.Sprintf("Interface %q not found.", interfaceID))
		return
	}

	emptyMembers, diags := types.ListValueFrom(ctx, types.StringType, []types.String{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := ConfigInterfaceBondResourceModel{
		InterfaceID: types.StringValue(interfaceID),
		Members:     emptyMembers,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigInterfaceVLANResource{}
	_ resource.ResourceWithImportState = &ConfigInterfaceVLANResource{}
)

func NewConfigInterfaceVLANResource() resource.Resource { return &ConfigInterfaceVLANResource{} }

type ConfigInterfaceVLANResource struct {
	client *loadmaster.Client
}

type ConfigInterfaceVLANResourceModel struct {
	InterfaceID types.String `tfsdk:"interface_id"`
	VLANID      types.Int64  `tfsdk:"vlan_id"`
}

func (r *ConfigInterfaceVLANResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_interface_vlan"
}

func (r *ConfigInterfaceVLANResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VLAN on a Kemp LoadMaster network interface.",
		Attributes: map[string]schema.Attribute{
			"interface_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Interface index (e.g. `\"1\"`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vlan_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** VLAN ID (1–4094). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ConfigInterfaceVLANResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigInterfaceVLANResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigInterfaceVLANResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddVLAN(ctx, data.InterfaceID.ValueString(), data.VLANID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error adding VLAN", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigInterfaceVLANResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigInterfaceVLANResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// The VLAN presence cannot be reliably verified via showiface alone without
	// knowing the VLAN sub-interface naming. Keep current state to avoid drift.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigInterfaceVLANResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All attributes are ForceNew — no in-place update path.
}

func (r *ConfigInterfaceVLANResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigInterfaceVLANResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVLAN(ctx, data.InterfaceID.ValueString(), data.VLANID.ValueInt64()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error removing VLAN", err.Error())
	}
}

func (r *ConfigInterfaceVLANResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID: "<interface_id>/<vlan_id>"
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format <interface_id>/<vlan_id>, got %q.", req.ID))
		return
	}

	var vlanID int64
	if _, err := fmt.Sscanf(parts[1], "%d", &vlanID); err != nil {
		resp.Diagnostics.AddError("Invalid VLAN ID in import ID", fmt.Sprintf("Could not parse VLAN ID from %q: %s.", parts[1], err))
		return
	}

	data := ConfigInterfaceVLANResourceModel{
		InterfaceID: types.StringValue(parts[0]),
		VLANID:      types.Int64Value(vlanID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

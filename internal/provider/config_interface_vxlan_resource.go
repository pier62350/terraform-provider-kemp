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
	_ resource.Resource                = &ConfigInterfaceVXLANResource{}
	_ resource.ResourceWithImportState = &ConfigInterfaceVXLANResource{}
)

func NewConfigInterfaceVXLANResource() resource.Resource { return &ConfigInterfaceVXLANResource{} }

type ConfigInterfaceVXLANResource struct {
	client *loadmaster.Client
}

type ConfigInterfaceVXLANResourceModel struct {
	InterfaceID types.String `tfsdk:"interface_id"`
	VNI         types.Int64  `tfsdk:"vni"`
	Remote      types.String `tfsdk:"remote"`
}

func (r *ConfigInterfaceVXLANResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_interface_vxlan"
}

func (r *ConfigInterfaceVXLANResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VXLAN tunnel on a Kemp LoadMaster network interface.",
		Attributes: map[string]schema.Attribute{
			"interface_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Interface index (e.g. `\"1\"`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vni": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** VXLAN Network Identifier.",
			},
			"remote": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Remote VXLAN tunnel endpoint IP address.",
			},
		},
	}
}

func (r *ConfigInterfaceVXLANResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigInterfaceVXLANResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigInterfaceVXLANResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddVXLAN(ctx, data.InterfaceID.ValueString(), data.VNI.ValueInt64(), data.Remote.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error adding VXLAN", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigInterfaceVXLANResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigInterfaceVXLANResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := r.client.ShowVXLAN(ctx, data.InterfaceID.ValueString())
	if err != nil || info == nil {
		// If not found or error, remove from state.
		if err != nil && !loadmaster.IsNotFound(err) {
			resp.Diagnostics.AddError("Error reading VXLAN", err.Error())
			return
		}
		resp.State.RemoveResource(ctx)
		return
	}

	data.VNI = types.Int64Value(info.Vni)
	data.Remote = types.StringValue(info.Remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigInterfaceVXLANResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigInterfaceVXLANResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ModifyVXLAN(ctx, data.InterfaceID.ValueString(), data.VNI.ValueInt64(), data.Remote.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating VXLAN", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigInterfaceVXLANResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigInterfaceVXLANResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVXLAN(ctx, data.InterfaceID.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error removing VXLAN", err.Error())
	}
}

func (r *ConfigInterfaceVXLANResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	interfaceID := req.ID
	info, err := r.client.ShowVXLAN(ctx, interfaceID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing VXLAN", err.Error())
		return
	}
	if info == nil {
		resp.Diagnostics.AddError("VXLAN not found", fmt.Sprintf("No VXLAN tunnel found on interface %q.", interfaceID))
		return
	}

	data := ConfigInterfaceVXLANResourceModel{
		InterfaceID: types.StringValue(interfaceID),
		VNI:         types.Int64Value(info.Vni),
		Remote:      types.StringValue(info.Remote),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

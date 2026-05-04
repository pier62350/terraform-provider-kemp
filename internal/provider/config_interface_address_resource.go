// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigInterfaceAddressResource{}
	_ resource.ResourceWithImportState = &ConfigInterfaceAddressResource{}
)

func NewConfigInterfaceAddressResource() resource.Resource { return &ConfigInterfaceAddressResource{} }

type ConfigInterfaceAddressResource struct {
	client *loadmaster.Client
}

type ConfigInterfaceAddressResourceModel struct {
	InterfaceID types.String `tfsdk:"interface_id"`
	Address     types.String `tfsdk:"address"`
}

func (r *ConfigInterfaceAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_interface_address"
}

func (r *ConfigInterfaceAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an additional IP address on a Kemp LoadMaster network interface.",
		Attributes: map[string]schema.Attribute{
			"interface_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Interface index (e.g. `\"1\"`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Additional IP address to assign to the interface. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ConfigInterfaceAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigInterfaceAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigInterfaceAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddInterfaceAddress(ctx, data.InterfaceID.ValueString(), data.Address.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error adding interface address", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigInterfaceAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigInterfaceAddressResourceModel
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

	addr := data.Address.ValueString()
	for _, a := range iface.AdditionalAddresses {
		if strings.EqualFold(strings.TrimSpace(a), addr) {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *ConfigInterfaceAddressResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All attributes are ForceNew — no in-place update path.
}

func (r *ConfigInterfaceAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigInterfaceAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteInterfaceAddress(ctx, data.InterfaceID.ValueString(), data.Address.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error removing interface address", err.Error())
	}
}

func (r *ConfigInterfaceAddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID: "<interface_id>/<address>"
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format <interface_id>/<address>, got %q.", req.ID))
		return
	}

	data := ConfigInterfaceAddressResourceModel{
		InterfaceID: types.StringValue(parts[0]),
		Address:     types.StringValue(parts[1]),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigPacketRoutingFilterResource{}
	_ resource.ResourceWithImportState = &ConfigPacketRoutingFilterResource{}
)

func NewConfigPacketRoutingFilterResource() resource.Resource {
	return &ConfigPacketRoutingFilterResource{}
}

type ConfigPacketRoutingFilterResource struct {
	client *loadmaster.Client
}

type ConfigPacketRoutingFilterResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Drop                types.Bool   `tfsdk:"drop"`
	RestrictToInterface types.Bool   `tfsdk:"restrict_to_interface"`
	IncludeWUI          types.Bool   `tfsdk:"include_wui"`
}

func (r *ConfigPacketRoutingFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_packet_routing_filter"
}

func (r *ConfigPacketRoutingFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the packet routing (ACL) filter settings on a Kemp LoadMaster.

This is a singleton resource. Import with: ` + "`terraform import kemp_config_packet_routing_filter.main loadmaster`",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `loadmaster`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Enable or disable the packet routing filter.",
			},
			"drop": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Optional. Drop blocked connections instead of rejecting them. Default: `false`.",
			},
			"restrict_to_interface": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Optional. Restrict IP access lists to the interface. Default: `false`.",
			},
			"include_wui": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Optional. Include the Web User Interface in IP access lists. Default: `false`.",
			},
		},
	}
}

func (r *ConfigPacketRoutingFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func modelToPacketFilter(data ConfigPacketRoutingFilterResourceModel) loadmaster.PacketRoutingFilter {
	return loadmaster.PacketRoutingFilter{
		Enabled:             data.Enabled.ValueBool(),
		Drop:                data.Drop.ValueBool(),
		RestrictToInterface: data.RestrictToInterface.ValueBool(),
		IncludeWUI:          data.IncludeWUI.ValueBool(),
	}
}

func writePacketFilterState(f *loadmaster.PacketRoutingFilter, data *ConfigPacketRoutingFilterResourceModel) {
	data.ID = types.StringValue("loadmaster")
	data.Enabled = types.BoolValue(f.Enabled)
	data.Drop = types.BoolValue(f.Drop)
	data.RestrictToInterface = types.BoolValue(f.RestrictToInterface)
	data.IncludeWUI = types.BoolValue(f.IncludeWUI)
}

func (r *ConfigPacketRoutingFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigPacketRoutingFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetPacketRoutingFilter(ctx, modelToPacketFilter(data)); err != nil {
		resp.Diagnostics.AddError("Error configuring packet routing filter", err.Error())
		return
	}

	f, err := r.client.GetPacketRoutingFilter(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading packet routing filter after create", err.Error())
		return
	}
	writePacketFilterState(f, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigPacketRoutingFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigPacketRoutingFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, err := r.client.GetPacketRoutingFilter(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading packet routing filter", err.Error())
		return
	}
	writePacketFilterState(f, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigPacketRoutingFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigPacketRoutingFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetPacketRoutingFilter(ctx, modelToPacketFilter(data)); err != nil {
		resp.Diagnostics.AddError("Error updating packet routing filter", err.Error())
		return
	}

	f, err := r.client.GetPacketRoutingFilter(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading packet routing filter after update", err.Error())
		return
	}
	writePacketFilterState(f, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigPacketRoutingFilterResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton — removing from state is sufficient; LoadMaster config is unchanged.
}

func (r *ConfigPacketRoutingFilterResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	f, err := r.client.GetPacketRoutingFilter(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing packet routing filter", err.Error())
		return
	}
	var data ConfigPacketRoutingFilterResourceModel
	writePacketFilterState(f, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

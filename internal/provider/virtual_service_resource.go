// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &VirtualServiceResource{}
	_ resource.ResourceWithImportState = &VirtualServiceResource{}
)

func NewVirtualServiceResource() resource.Resource {
	return &VirtualServiceResource{}
}

type VirtualServiceResource struct {
	client *loadmaster.Client
}

type VirtualServiceResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Address  types.String `tfsdk:"address"`
	Port     types.String `tfsdk:"port"`
	Protocol types.String `tfsdk:"protocol"`
	Type     types.String `tfsdk:"type"`
	Nickname types.String `tfsdk:"nickname"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

func (r *VirtualServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_service"
}

func (r *VirtualServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kemp LoadMaster virtual service (VS).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "LoadMaster `Index` of the virtual service.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "IP address of an interface attached to the LoadMaster.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Listening port of the virtual service.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Layer-4 protocol: `tcp` or `udp`.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "VS type — one of `gen`, `http`, `http2`, `ts`, `tls`, `log`.",
				Optional:            true,
				Computed:            true,
			},
			"nickname": schema.StringAttribute{
				MarkdownDescription: "Friendly name for the virtual service.",
				Optional:            true,
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the virtual service is enabled.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *VirtualServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*loadmaster.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *loadmaster.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *VirtualServiceResource) paramsFromModel(m VirtualServiceResourceModel) loadmaster.VirtualServiceParams {
	p := loadmaster.VirtualServiceParams{
		NickName: m.Nickname.ValueString(),
		VSType:   m.Type.ValueString(),
	}
	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		p.Enable = boolPtr(m.Enabled.ValueBool())
	}
	return p
}

func (r *VirtualServiceResource) writeState(vs *loadmaster.VirtualService, m *VirtualServiceResourceModel) {
	m.Id = types.StringValue(strconv.Itoa(int(vs.Index)))
	m.Address = types.StringValue(vs.Address)
	m.Port = types.StringValue(vs.Port)
	m.Protocol = types.StringValue(vs.Protocol)
	m.Type = types.StringValue(vs.VSType)
	m.Nickname = types.StringValue(vs.NickName)
	if vs.Enable != nil {
		m.Enabled = types.BoolValue(*vs.Enable)
	} else {
		m.Enabled = types.BoolValue(false)
	}
}

func (r *VirtualServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.AddVirtualService(ctx, data.Address.ValueString(), data.Port.ValueString(), data.Protocol.ValueString(), r.paramsFromModel(data))
	if err != nil {
		resp.Diagnostics.AddError("Error creating virtual service", err.Error())
		return
	}
	tflog.Trace(ctx, "created virtual service", map[string]any{"index": vs.Index})

	r.writeState(vs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.ShowVirtualService(ctx, data.Id.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading virtual service", err.Error())
		return
	}

	r.writeState(vs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.ModifyVirtualService(ctx, data.Id.ValueString(), r.paramsFromModel(data))
	if err != nil {
		resp.Diagnostics.AddError("Error updating virtual service", err.Error())
		return
	}

	r.writeState(vs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVirtualService(ctx, data.Id.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting virtual service", err.Error())
	}
}

func (r *VirtualServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	vs, err := r.client.ShowVirtualService(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing virtual service", err.Error())
		return
	}

	var data VirtualServiceResourceModel
	r.writeState(vs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

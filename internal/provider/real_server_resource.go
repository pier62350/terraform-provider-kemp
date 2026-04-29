// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &RealServerResource{}
	_ resource.ResourceWithImportState = &RealServerResource{}
)

func NewRealServerResource() resource.Resource { return &RealServerResource{} }

type RealServerResource struct {
	client *loadmaster.Client
}

type RealServerResourceModel struct {
	Id               types.Int32  `tfsdk:"id"`
	VirtualServiceId types.String `tfsdk:"virtual_service_id"`
	Address          types.String `tfsdk:"address"`
	Port             types.String `tfsdk:"port"`
	Weight           types.Int32  `tfsdk:"weight"`
	Forward          types.String `tfsdk:"forward"`
	Enable           types.Bool   `tfsdk:"enable"`
	Limit            types.Int32  `tfsdk:"limit"`
	Critical         types.Bool   `tfsdk:"critical"`
	Follow           types.Int32  `tfsdk:"follow"`
	DnsName          types.String `tfsdk:"dns_name"`
}

func (r *RealServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_real_server"
}

func (r *RealServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a real server (backend) attached to a Kemp virtual service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "LoadMaster `RsIndex` of the real server.",
				PlanModifiers:       []planmodifier.Int32{int32planmodifier.UseStateForUnknown()},
			},
			"virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Index of the parent virtual service.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IP address of the real server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"port": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Port on the real server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"weight":   schema.Int32Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Load-balancing weight relative to other real servers in the pool. Default: `1000`."},
			"forward":  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Forwarding method: `nat` (default), `route`, or `transparent`."},
			"enable":   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Whether the real server is enabled. Default: `true`."},
			"limit":    schema.Int32Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Maximum concurrent connections to this server. Default: `0` (unlimited)."},
			"critical": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. If true, the entire VS is marked down when this server fails. Default: `false`."},
			"follow":   schema.Int32Attribute{Optional: true, Computed: true, MarkdownDescription: "Optional. `RsIndex` of another real server to mirror health status from. Default: `0` (disabled)."},
			"dns_name": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. DNS hostname override for this real server. Empty string disables override."},
		},
	}
}

func (r *RealServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RealServerResource) paramsFromModel(m RealServerResourceModel) loadmaster.RealServerParams {
	p := loadmaster.RealServerParams{
		Weight:  m.Weight.ValueInt32(),
		Forward: m.Forward.ValueString(),
		Limit:   m.Limit.ValueInt32(),
		Follow:  m.Follow.ValueInt32(),
	}
	if !m.Enable.IsNull() && !m.Enable.IsUnknown() {
		p.Enable = boolPtr(m.Enable.ValueBool())
	}
	if !m.Critical.IsNull() && !m.Critical.IsUnknown() {
		p.Critical = boolPtr(m.Critical.ValueBool())
	}
	return p
}

func (r *RealServerResource) writeState(rs *loadmaster.RealServer, m *RealServerResourceModel) {
	m.Id = types.Int32Value(rs.RsIndex)
	m.VirtualServiceId = types.StringValue(strconv.Itoa(int(rs.VSIndex)))
	m.Address = types.StringValue(rs.Address)
	m.Port = types.StringValue(strconv.Itoa(int(rs.Port)))
	m.Weight = types.Int32Value(rs.Weight)
	m.Forward = types.StringValue(rs.Forward)
	m.Limit = types.Int32Value(rs.Limit)
	m.Follow = types.Int32Value(rs.Follow)
	m.DnsName = types.StringValue(rs.DnsName)
	if rs.Enable != nil {
		m.Enable = types.BoolValue(*rs.Enable)
	} else {
		m.Enable = types.BoolValue(false)
	}
	if rs.Critical != nil {
		m.Critical = types.BoolValue(*rs.Critical)
	} else {
		m.Critical = types.BoolValue(false)
	}
}

func (r *RealServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RealServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rs, err := r.client.AddRealServer(ctx, data.VirtualServiceId.ValueString(), data.Address.ValueString(), data.Port.ValueString(), r.paramsFromModel(data))
	if err != nil {
		resp.Diagnostics.AddError("Error creating real server", err.Error())
		return
	}
	r.writeState(rs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RealServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RealServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rs, err := r.client.ShowRealServer(ctx, data.VirtualServiceId.ValueString(), strconv.Itoa(int(data.Id.ValueInt32())))
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading real server", err.Error())
		return
	}

	r.writeState(rs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RealServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RealServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rs, err := r.client.ModifyRealServer(ctx, data.VirtualServiceId.ValueString(), strconv.Itoa(int(data.Id.ValueInt32())), r.paramsFromModel(data))
	if err != nil {
		resp.Diagnostics.AddError("Error updating real server", err.Error())
		return
	}

	r.writeState(rs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RealServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RealServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRealServer(ctx, data.VirtualServiceId.ValueString(), strconv.Itoa(int(data.Id.ValueInt32()))); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting real server", err.Error())
	}
}

// ImportState accepts an ID of the form "<vs_index>/<rs_index>".
func (r *RealServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(`expected "<vs_index>/<rs_index>", got %q`, req.ID))
		return
	}

	rs, err := r.client.ShowRealServer(ctx, parts[0], parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Error importing real server", err.Error())
		return
	}

	var data RealServerResourceModel
	r.writeState(rs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &SubVirtualServiceResource{}
	_ resource.ResourceWithImportState = &SubVirtualServiceResource{}
)

func NewSubVirtualServiceResource() resource.Resource { return &SubVirtualServiceResource{} }

type SubVirtualServiceResource struct {
	client *loadmaster.Client
}

type SubVirtualServiceResourceModel struct {
	Id       types.String `tfsdk:"id"`
	ParentId types.String `tfsdk:"parent_id"`
	Nickname types.String `tfsdk:"nickname"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Type     types.String `tfsdk:"type"`
}

func (r *SubVirtualServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sub_virtual_service"
}

func (r *SubVirtualServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Sub-Virtual Service (SubVS) under a parent ` + "`kemp_virtual_service`" + `.

A SubVS shares the parent's listening address/port/protocol and adds L7 routing on top — typically used to dispatch traffic by host/path to different real-server pools. Sub-VS creation goes through the parent's ` + "`modvs`" + ` with ` + "`createsubvs`" + `; thereafter the SubVS has its own Index used for CRUD.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "LoadMaster `Index` assigned to the SubVS.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"parent_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "`Index` of the parent virtual service this SubVS attaches to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"nickname": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Friendly name shown in the WUI.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the SubVS is enabled.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "VS type (e.g. `http`, `gen`, `tls`). Inherits from parent if unset.",
			},
		},
	}
}

func (r *SubVirtualServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubVirtualServiceResource) writeState(vs *loadmaster.VirtualService, m *SubVirtualServiceResourceModel) {
	m.Id = types.StringValue(strconv.Itoa(int(vs.Index)))
	m.Type = types.StringValue(vs.VSType)
	m.Nickname = types.StringValue(vs.NickName)
	if vs.Enable != nil {
		m.Enabled = types.BoolValue(*vs.Enable)
	} else {
		m.Enabled = types.BoolValue(false)
	}
}

func (r *SubVirtualServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubVirtualServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.CreateSubVS(ctx, data.ParentId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating sub-virtual service", err.Error())
		return
	}

	// After creation, apply any user-set fields via modvs on the new SubVS.
	if needsFollowUpModify(data) {
		params := loadmaster.VirtualServiceParams{
			NickName: data.Nickname.ValueString(),
			VSType:   data.Type.ValueString(),
		}
		if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
			params.Enable = boolPtr(data.Enabled.ValueBool())
		}
		updated, mErr := r.client.ModifyVirtualService(ctx, strconv.Itoa(int(vs.Index)), params)
		if mErr != nil {
			resp.Diagnostics.AddError("Error applying sub-virtual service settings post-create", mErr.Error())
			return
		}
		vs = updated
	}

	r.writeState(vs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func needsFollowUpModify(m SubVirtualServiceResourceModel) bool {
	if !m.Nickname.IsNull() && !m.Nickname.IsUnknown() && m.Nickname.ValueString() != "" {
		return true
	}
	if !m.Type.IsNull() && !m.Type.IsUnknown() && m.Type.ValueString() != "" {
		return true
	}
	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		return true
	}
	return false
}

func (r *SubVirtualServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubVirtualServiceResourceModel
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
		resp.Diagnostics.AddError("Error reading sub-virtual service", err.Error())
		return
	}

	r.writeState(vs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubVirtualServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubVirtualServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := loadmaster.VirtualServiceParams{
		NickName: data.Nickname.ValueString(),
		VSType:   data.Type.ValueString(),
	}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		params.Enable = boolPtr(data.Enabled.ValueBool())
	}

	vs, err := r.client.ModifyVirtualService(ctx, data.Id.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating sub-virtual service", err.Error())
		return
	}

	r.writeState(vs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubVirtualServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubVirtualServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVirtualService(ctx, data.Id.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting sub-virtual service", err.Error())
	}
}

// ImportState accepts "<parent_id>/<subvs_id>" so the parent reference is
// preserved in state without needing an additional API call.
func (r *SubVirtualServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(`expected "<parent_id>/<subvs_id>", got %q`, req.ID))
		return
	}

	vs, err := r.client.ShowVirtualService(ctx, parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Error importing sub-virtual service", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_id"), parts[0])...)
	data := SubVirtualServiceResourceModel{ParentId: types.StringValue(parts[0])}
	r.writeState(vs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

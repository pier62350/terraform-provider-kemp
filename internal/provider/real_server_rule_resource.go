// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
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
	_ resource.Resource                = &RealServerRuleResource{}
	_ resource.ResourceWithImportState = &RealServerRuleResource{}
)

func NewRealServerRuleResource() resource.Resource { return &RealServerRuleResource{} }

type RealServerRuleResource struct {
	client *loadmaster.Client
}

type RealServerRuleModel struct {
	VirtualServiceId types.String `tfsdk:"virtual_service_id"`
	RealServerId     types.String `tfsdk:"real_server_id"`
	VSPort           types.String `tfsdk:"vs_port"`
	VSProtocol       types.String `tfsdk:"vs_protocol"`
	RSAddress        types.String `tfsdk:"rs_address"`
	RSPort           types.String `tfsdk:"rs_port"`
	Rule             types.String `tfsdk:"rule"`
}

func (r *RealServerRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_real_server_rule"
}

func (r *RealServerRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches an existing system-level rule to a real server (backend pool member) on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Numeric index of the parent virtual service. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"real_server_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Numeric index of the real server (without the `!` prefix). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vs_port": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Port of the parent virtual service (e.g. `80`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vs_protocol": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Protocol of the parent virtual service (`tcp` or `udp`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"rs_address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IP address of the real server. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"rs_port": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Port of the real server (e.g. `8080`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"rule": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the system rule to attach. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *RealServerRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RealServerRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RealServerRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddRealServerRule(ctx,
		data.VirtualServiceId.ValueString(),
		data.VSPort.ValueString(),
		data.VSProtocol.ValueString(),
		data.RSAddress.ValueString(),
		data.RSPort.ValueString(),
		data.Rule.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error attaching rule to real server", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RealServerRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RealServerRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	attached, err := r.client.RealServerHasRule(ctx,
		data.VirtualServiceId.ValueString(),
		data.RealServerId.ValueString(),
		data.Rule.ValueString(),
	)
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading real server rule attachment", err.Error())
		return
	}
	if !attached {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RealServerRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are ForceNew; framework should never call Update.
	resp.Diagnostics.AddError("Update not supported", "kemp_real_server_rule has no in-place updatable attributes; changes trigger replacement.")
}

func (r *RealServerRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RealServerRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRealServerRule(ctx,
		data.VirtualServiceId.ValueString(),
		data.VSPort.ValueString(),
		data.VSProtocol.ValueString(),
		data.RSAddress.ValueString(),
		data.RSPort.ValueString(),
		data.Rule.ValueString(),
	); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error detaching rule from real server", err.Error())
	}
}

// ImportState accepts "<virtual_service_id>/<real_server_id>/<vs_port>/<vs_protocol>/<rs_address>/<rs_port>/<rule_name>".
func (r *RealServerRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 7)
	if len(parts) != 7 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf(`expected "<virtual_service_id>/<real_server_id>/<vs_port>/<vs_protocol>/<rs_address>/<rs_port>/<rule_name>", got %q`, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_service_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("real_server_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vs_port"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vs_protocol"), parts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rs_address"), parts[4])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rs_port"), parts[5])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule"), parts[6])...)
}

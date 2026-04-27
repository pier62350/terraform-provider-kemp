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
	_ resource.Resource                = &VirtualServiceRuleResource{}
	_ resource.ResourceWithImportState = &VirtualServiceRuleResource{}
)

func NewVirtualServiceRuleResource() resource.Resource { return &VirtualServiceRuleResource{} }

type VirtualServiceRuleResource struct {
	client *loadmaster.Client
}

type VirtualServiceRuleModel struct {
	VirtualServiceId types.String `tfsdk:"virtual_service_id"`
	Rule             types.String `tfsdk:"rule"`
	Direction        types.String `tfsdk:"direction"`
}

func (r *VirtualServiceRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_service_rule"
}

func (r *VirtualServiceRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches an existing system-level rule (created via `kemp_add_header_rule`, `kemp_match_content_rule`, etc.) to a virtual service in a specific direction.",
		Attributes: map[string]schema.Attribute{
			"virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Index of the virtual service to attach the rule to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"rule": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the system rule to attach.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"direction": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "When the rule fires: `request` (incoming request, default for content/header rules), `response` (outgoing headers), `responsebody` (outgoing body), `pre` (pre-processing).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *VirtualServiceRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtualServiceRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtualServiceRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddVSRule(ctx, data.VirtualServiceId.ValueString(), data.Rule.ValueString(), data.Direction.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error attaching rule to virtual service", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtualServiceRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	attached, err := r.client.VSHasRule(ctx, data.VirtualServiceId.ValueString(), data.Rule.ValueString(), data.Direction.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading virtual service rule attachment", err.Error())
		return
	}
	if !attached {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are ForceNew; framework should never call Update.
	resp.Diagnostics.AddError("Update not supported", "kemp_virtual_service_rule has no in-place updatable attributes; changes trigger replacement.")
}

func (r *VirtualServiceRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtualServiceRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVSRule(ctx, data.VirtualServiceId.ValueString(), data.Rule.ValueString(), data.Direction.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error detaching rule", err.Error())
	}
}

// ImportState accepts "<virtual_service_id>/<direction>/<rule_name>".
func (r *VirtualServiceRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(`expected "<virtual_service_id>/<direction>/<rule_name>", got %q`, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_service_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("direction"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule"), parts[2])...)
}

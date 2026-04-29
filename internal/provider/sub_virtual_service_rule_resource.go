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
	_ resource.Resource                = &SubVirtualServiceRuleResource{}
	_ resource.ResourceWithImportState = &SubVirtualServiceRuleResource{}
)

func NewSubVirtualServiceRuleResource() resource.Resource {
	return &SubVirtualServiceRuleResource{}
}

type SubVirtualServiceRuleResource struct {
	client *loadmaster.Client
}

type SubVirtualServiceRuleModel struct {
	ParentVirtualServiceId types.String `tfsdk:"parent_virtual_service_id"`
	SubVirtualServiceId    types.String `tfsdk:"sub_virtual_service_id"`
	Rule                   types.String `tfsdk:"rule"`
}

func (r *SubVirtualServiceRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sub_virtual_service_rule"
}

func (r *SubVirtualServiceRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Attaches a content-switching rule to a SubVS.

Content switching routes incoming requests to different SubVSes based on URL, headers, or other match criteria. Attach a rule (created via ` + "`kemp_match_content_rule`" + ` or similar) to a SubVS: the parent VS evaluates the rule to decide whether a request should go to that SubVS.

Uses the LoadMaster ` + "`addrsrule`" + ` command internally.`,
		Attributes: map[string]schema.Attribute{
			"parent_virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "`Index` of the parent virtual service that owns this SubVS.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"sub_virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "`Index` of the SubVS to attach the rule to (from `kemp_sub_virtual_service.*.id`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"rule": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the system rule to attach (e.g. from `kemp_match_content_rule.*.name`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *SubVirtualServiceRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubVirtualServiceRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubVirtualServiceRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddSubVSRule(ctx, data.ParentVirtualServiceId.ValueString(), data.SubVirtualServiceId.ValueString(), data.Rule.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error attaching rule to SubVS", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubVirtualServiceRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubVirtualServiceRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Verify the SubVS still exists; if not, remove from state.
	_, err := r.client.ShowVirtualService(ctx, data.SubVirtualServiceId.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SubVS for rule attachment", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubVirtualServiceRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "kemp_sub_virtual_service_rule has no in-place updatable attributes; changes trigger replacement.")
}

func (r *SubVirtualServiceRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubVirtualServiceRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSubVSRule(ctx, data.ParentVirtualServiceId.ValueString(), data.SubVirtualServiceId.ValueString(), data.Rule.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error detaching rule from SubVS", err.Error())
	}
}

// ImportState accepts "<parent_vs_id>/<sub_vs_id>/<rule_name>".
func (r *SubVirtualServiceRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(`expected "<parent_vs_id>/<sub_vs_id>/<rule_name>", got %q`, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_virtual_service_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sub_virtual_service_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule"), parts[2])...)
}

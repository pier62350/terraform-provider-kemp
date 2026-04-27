// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &DeleteHeaderRuleResource{}
	_ resource.ResourceWithImportState = &DeleteHeaderRuleResource{}
)

func NewDeleteHeaderRuleResource() resource.Resource { return &DeleteHeaderRuleResource{} }

type DeleteHeaderRuleResource struct {
	client *loadmaster.Client
}

type DeleteHeaderRuleModel struct {
	Name         types.String `tfsdk:"name"`
	Pattern      types.String `tfsdk:"pattern"`
	OnlyOnFlag   types.Int64  `tfsdk:"only_on_flag"`
	OnlyOnNoFlag types.Int64  `tfsdk:"only_on_no_flag"`
}

func (r *DeleteHeaderRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delete_header_rule"
}

func (r *DeleteHeaderRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "System-level rule that strips request headers matching `pattern`. Attach to a virtual service via `kemp_virtual_service_rule`.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"pattern": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Header-name pattern to delete (regex).",
			},
			"only_on_flag":    schema.Int64Attribute{Optional: true},
			"only_on_no_flag": schema.Int64Attribute{Optional: true},
		},
	}
}

func (r *DeleteHeaderRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (m DeleteHeaderRuleModel) toParams() loadmaster.RuleParams {
	p := loadmaster.RuleParams{Pattern: m.Pattern.ValueString()}
	if !m.OnlyOnFlag.IsNull() && !m.OnlyOnFlag.IsUnknown() {
		v := int32(m.OnlyOnFlag.ValueInt64())
		p.OnlyOnFlag = &v
	}
	if !m.OnlyOnNoFlag.IsNull() && !m.OnlyOnNoFlag.IsUnknown() {
		v := int32(m.OnlyOnNoFlag.ValueInt64())
		p.OnlyOnNoFlag = &v
	}
	return p
}

func (r *DeleteHeaderRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeleteHeaderRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p := data.toParams()
	p.Type = loadmaster.RuleTypeDeleteHeader
	if err := r.client.AddRule(ctx, data.Name.ValueString(), p); err != nil {
		resp.Diagnostics.AddError("Error creating delete_header_rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeleteHeaderRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeleteHeaderRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule, err := r.client.FindDeleteHeaderRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading delete_header_rule", err.Error())
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Pattern = types.StringValue(rule.Pattern)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeleteHeaderRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeleteHeaderRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ModifyRule(ctx, data.Name.ValueString(), data.toParams()); err != nil {
		resp.Diagnostics.AddError("Error updating delete_header_rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeleteHeaderRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeleteHeaderRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRule(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting delete_header_rule", err.Error())
	}
}

func (r *DeleteHeaderRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rule, err := r.client.FindDeleteHeaderRule(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing delete_header_rule", err.Error())
		return
	}
	if rule == nil {
		resp.Diagnostics.AddError("Rule not found", fmt.Sprintf("no DeleteHeaderRule named %q on the LoadMaster", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pattern"), rule.Pattern)...)
}

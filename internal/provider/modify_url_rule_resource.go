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
	_ resource.Resource                = &ModifyURLRuleResource{}
	_ resource.ResourceWithImportState = &ModifyURLRuleResource{}
)

func NewModifyURLRuleResource() resource.Resource { return &ModifyURLRuleResource{} }

type ModifyURLRuleResource struct {
	client *loadmaster.Client
}

type ModifyURLRuleModel struct {
	Name         types.String `tfsdk:"name"`
	Pattern      types.String `tfsdk:"pattern"`
	Replacement  types.String `tfsdk:"replacement"`
	OnlyOnFlag   types.Int64  `tfsdk:"only_on_flag"`
	OnlyOnNoFlag types.Int64  `tfsdk:"only_on_no_flag"`
}

func (r *ModifyURLRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_modify_url_rule"
}

func (r *ModifyURLRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "System-level URL rewrite rule. Attach via `kemp_virtual_service_rule`.",
		Attributes: map[string]schema.Attribute{
			"name":            schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"pattern":         schema.StringAttribute{Required: true, MarkdownDescription: "URL pattern to match."},
			"replacement":     schema.StringAttribute{Required: true, MarkdownDescription: "Replacement URL/path."},
			"only_on_flag":    schema.Int64Attribute{Optional: true},
			"only_on_no_flag": schema.Int64Attribute{Optional: true},
		},
	}
}

func (r *ModifyURLRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (m ModifyURLRuleModel) toParams() loadmaster.RuleParams {
	p := loadmaster.RuleParams{Pattern: m.Pattern.ValueString(), Replacement: m.Replacement.ValueString()}
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

func (r *ModifyURLRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ModifyURLRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p := data.toParams()
	p.Type = loadmaster.RuleTypeModifyURL
	if err := r.client.AddRule(ctx, data.Name.ValueString(), p); err != nil {
		resp.Diagnostics.AddError("Error creating modify_url_rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModifyURLRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ModifyURLRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule, err := r.client.FindModifyURLRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading modify_url_rule", err.Error())
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Pattern = types.StringValue(rule.Pattern)
	data.Replacement = types.StringValue(rule.Replacement)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModifyURLRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ModifyURLRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ModifyRule(ctx, data.Name.ValueString(), data.toParams()); err != nil {
		resp.Diagnostics.AddError("Error updating modify_url_rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModifyURLRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ModifyURLRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRule(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting modify_url_rule", err.Error())
	}
}

func (r *ModifyURLRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rule, err := r.client.FindModifyURLRule(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing modify_url_rule", err.Error())
		return
	}
	if rule == nil {
		resp.Diagnostics.AddError("Rule not found", fmt.Sprintf("no ModifyURLRule named %q on the LoadMaster", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pattern"), rule.Pattern)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("replacement"), rule.Replacement)...)
}

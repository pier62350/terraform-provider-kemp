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
	_ resource.Resource                = &AddHeaderRuleResource{}
	_ resource.ResourceWithImportState = &AddHeaderRuleResource{}
)

func NewAddHeaderRuleResource() resource.Resource { return &AddHeaderRuleResource{} }

type AddHeaderRuleResource struct {
	client *loadmaster.Client
}

type AddHeaderRuleModel struct {
	Name         types.String `tfsdk:"name"`
	Header       types.String `tfsdk:"header"`
	Replacement  types.String `tfsdk:"replacement"`
	OnlyOnFlag   types.Int64  `tfsdk:"only_on_flag"`
	OnlyOnNoFlag types.Int64  `tfsdk:"only_on_no_flag"`
}

func (r *AddHeaderRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_add_header_rule"
}

func (r *AddHeaderRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "System-level rule that injects a request header. Attach to a virtual service via `kemp_virtual_service_rule`.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Unique rule name. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"header": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Header field name to inject.",
			},
			"replacement": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Header value to set (max 255 chars).",
			},
			"only_on_flag": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional. Run only if rule-chain flag (1–9) is set.",
			},
			"only_on_no_flag": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional. Run only if rule-chain flag (1–9) is NOT set.",
			},
		},
	}
}

func (r *AddHeaderRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (m AddHeaderRuleModel) toParams() loadmaster.RuleParams {
	p := loadmaster.RuleParams{
		Header:      m.Header.ValueString(),
		Replacement: m.Replacement.ValueString(),
	}
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

func (r *AddHeaderRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AddHeaderRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p := data.toParams()
	p.Type = loadmaster.RuleTypeAddHeader
	if err := r.client.AddRule(ctx, data.Name.ValueString(), p); err != nil {
		resp.Diagnostics.AddError("Error creating add_header_rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AddHeaderRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AddHeaderRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule, err := r.client.FindAddHeaderRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading add_header_rule", err.Error())
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Header = types.StringValue(rule.Header)
	data.Replacement = types.StringValue(rule.HeaderValue)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AddHeaderRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AddHeaderRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ModifyRule(ctx, data.Name.ValueString(), data.toParams()); err != nil {
		resp.Diagnostics.AddError("Error updating add_header_rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AddHeaderRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AddHeaderRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRule(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting add_header_rule", err.Error())
	}
}

func (r *AddHeaderRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rule, err := r.client.FindAddHeaderRule(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing add_header_rule", err.Error())
		return
	}
	if rule == nil {
		resp.Diagnostics.AddError("Rule not found", fmt.Sprintf("no AddHeaderRule named %q on the LoadMaster", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("header"), rule.Header)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("replacement"), rule.HeaderValue)...)
}

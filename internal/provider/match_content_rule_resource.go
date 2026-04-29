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
	_ resource.Resource                = &MatchContentRuleResource{}
	_ resource.ResourceWithImportState = &MatchContentRuleResource{}
)

func NewMatchContentRuleResource() resource.Resource { return &MatchContentRuleResource{} }

type MatchContentRuleResource struct {
	client *loadmaster.Client
}

type MatchContentRuleModel struct {
	Name            types.String `tfsdk:"name"`
	Pattern         types.String `tfsdk:"pattern"`
	MatchType       types.String `tfsdk:"match_type"`
	Header          types.String `tfsdk:"header"`
	IncludeHost     types.Bool   `tfsdk:"include_host"`
	IgnoreCase      types.Bool   `tfsdk:"ignore_case"`
	Negate          types.Bool   `tfsdk:"negate"`
	IncludeQuery    types.Bool   `tfsdk:"include_query"`
	MustFail        types.Bool   `tfsdk:"must_fail"`
	SetOnMatch      types.Int64  `tfsdk:"set_on_match"`
	OnlyOnFlag      types.Int64  `tfsdk:"only_on_flag"`
	OnlyOnNoFlag    types.Int64  `tfsdk:"only_on_no_flag"`
}

func (r *MatchContentRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_match_content_rule"
}

func (r *MatchContentRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "System-level content-match rule (regex/prefix/postfix) used to drive routing or chain other rules. Attach via `kemp_virtual_service_rule`.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Unique rule name. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"pattern":         schema.StringAttribute{Required: true, MarkdownDescription: "**Required.** Pattern to match against the URL (default scope) or the field set by `header`."},
			"match_type":      schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Match algorithm: `regex` (default), `prefix`, or `postfix`."},
			"header":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Header field to match against instead of the URL. Use `body` to match the request body."},
			"include_host":    schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Prepend the hostname to the request URI before matching. Default: `false`."},
			"ignore_case":     schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Case-insensitive matching. Default: `false`."},
			"negate":          schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Invert the sense of the match. Default: `false`."},
			"include_query":   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. Append the query string to the URI before matching. Default: `false`."},
			"must_fail":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional. If matched, abort the connection. Default: `false`."},
			"set_on_match":    schema.Int64Attribute{Optional: true, MarkdownDescription: "Optional. Set rule-chain flag (0–9) when matched."},
			"only_on_flag":    schema.Int64Attribute{Optional: true, MarkdownDescription: "Optional. Run only if rule-chain flag (1–9) is set."},
			"only_on_no_flag": schema.Int64Attribute{Optional: true, MarkdownDescription: "Optional. Run only if rule-chain flag (1–9) is NOT set."},
		},
	}
}

func (r *MatchContentRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (m MatchContentRuleModel) toParams() loadmaster.RuleParams {
	p := loadmaster.RuleParams{
		Pattern:   m.Pattern.ValueString(),
		MatchType: m.MatchType.ValueString(),
		Header:    m.Header.ValueString(),
	}
	if !m.IncludeHost.IsNull() && !m.IncludeHost.IsUnknown() {
		p.IncHost = boolPtr(m.IncludeHost.ValueBool())
	}
	if !m.IgnoreCase.IsNull() && !m.IgnoreCase.IsUnknown() {
		p.NoCase = boolPtr(m.IgnoreCase.ValueBool())
	}
	if !m.Negate.IsNull() && !m.Negate.IsUnknown() {
		p.Negate = boolPtr(m.Negate.ValueBool())
	}
	if !m.IncludeQuery.IsNull() && !m.IncludeQuery.IsUnknown() {
		p.IncQuery = boolPtr(m.IncludeQuery.ValueBool())
	}
	if !m.MustFail.IsNull() && !m.MustFail.IsUnknown() {
		p.MustFail = boolPtr(m.MustFail.ValueBool())
	}
	if !m.SetOnMatch.IsNull() && !m.SetOnMatch.IsUnknown() {
		v := int32(m.SetOnMatch.ValueInt64())
		p.SetOnMatch = &v
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

func (r *MatchContentRuleResource) writeState(rule *loadmaster.MatchContentRule, m *MatchContentRuleModel) {
	m.Pattern = types.StringValue(rule.Pattern)
	m.MatchType = types.StringValue(rule.MatchType)
	m.Header = types.StringValue(rule.Header)
	m.IncludeHost = types.BoolValue(rule.AddHost)
	m.IgnoreCase = types.BoolValue(rule.CaseIndependent)
	m.Negate = types.BoolValue(rule.Negate)
	m.IncludeQuery = types.BoolValue(rule.IncludeQuery)
	m.MustFail = types.BoolValue(rule.MustFail)
}

func (r *MatchContentRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MatchContentRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p := data.toParams()
	p.Type = loadmaster.RuleTypeMatchContent
	if err := r.client.AddRule(ctx, data.Name.ValueString(), p); err != nil {
		resp.Diagnostics.AddError("Error creating match_content_rule", err.Error())
		return
	}
	rule, err := r.client.FindMatchContentRule(ctx, data.Name.ValueString())
	if err != nil || rule == nil {
		resp.Diagnostics.AddError("Error reading rule after create", fmt.Sprintf("err=%v rule=%v", err, rule))
		return
	}
	r.writeState(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MatchContentRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MatchContentRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule, err := r.client.FindMatchContentRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading match_content_rule", err.Error())
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	r.writeState(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MatchContentRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MatchContentRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ModifyRule(ctx, data.Name.ValueString(), data.toParams()); err != nil {
		resp.Diagnostics.AddError("Error updating match_content_rule", err.Error())
		return
	}
	rule, err := r.client.FindMatchContentRule(ctx, data.Name.ValueString())
	if err == nil && rule != nil {
		r.writeState(rule, &data)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MatchContentRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MatchContentRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRule(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting match_content_rule", err.Error())
	}
}

func (r *MatchContentRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rule, err := r.client.FindMatchContentRule(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing match_content_rule", err.Error())
		return
	}
	if rule == nil {
		resp.Diagnostics.AddError("Rule not found", fmt.Sprintf("no MatchContentRule named %q on the LoadMaster", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	data := MatchContentRuleModel{Name: types.StringValue(req.ID)}
	r.writeState(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigURLLimitRuleResource{}
	_ resource.ResourceWithImportState = &ConfigURLLimitRuleResource{}
)

func NewConfigURLLimitRuleResource() resource.Resource {
	return &ConfigURLLimitRuleResource{}
}

type ConfigURLLimitRuleResource struct {
	client *loadmaster.Client
}

type ConfigURLLimitRuleModel struct {
	Name    types.String `tfsdk:"name"`
	Pattern types.String `tfsdk:"pattern"`
	Limit   types.Int64  `tfsdk:"limit"`
	Match   types.Int64  `tfsdk:"match"`
}

func (r *ConfigURLLimitRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_url_limit_rule"
}

func (r *ConfigURLLimitRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a global URL-based rate limiting rule on a Kemp LoadMaster. Rules match incoming request URLs and enforce a per-rule request rate limit.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Unique rule name. Acts as the resource ID. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"pattern": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** URL pattern to match (e.g. `=/test/a.html` for exact match, `/api/` for prefix).",
			},
			"limit": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** Maximum requests per second allowed for URLs matching this rule.",
			},
			"match": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Optional. URL match type: `0` = exact (default), `1` = prefix, `2` = regex.",
			},
		},
	}
}

func (r *ConfigURLLimitRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigURLLimitRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigURLLimitRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddURLLimitRule(ctx, data.Name.ValueString(), data.Pattern.ValueString(), data.Limit.ValueInt64(), data.Match.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error creating URL limit rule", err.Error())
		return
	}

	rule, err := r.client.FindURLLimitRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading URL limit rule after create", err.Error())
		return
	}
	if rule != nil {
		data.Pattern = types.StringValue(rule.Pattern)
		data.Limit = types.Int64Value(rule.Limit)
		data.Match = types.Int64Value(rule.Match)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigURLLimitRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigURLLimitRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.FindURLLimitRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading URL limit rule", err.Error())
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Pattern = types.StringValue(rule.Pattern)
	data.Limit = types.Int64Value(rule.Limit)
	data.Match = types.Int64Value(rule.Match)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigURLLimitRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigURLLimitRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ModifyURLLimitRule(ctx, data.Name.ValueString(), data.Pattern.ValueString(), data.Limit.ValueInt64(), data.Match.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error updating URL limit rule", err.Error())
		return
	}

	rule, err := r.client.FindURLLimitRule(ctx, data.Name.ValueString())
	if err == nil && rule != nil {
		data.Pattern = types.StringValue(rule.Pattern)
		data.Limit = types.Int64Value(rule.Limit)
		data.Match = types.Int64Value(rule.Match)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigURLLimitRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigURLLimitRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteURLLimitRule(ctx, data.Name.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting URL limit rule", err.Error())
	}
}

func (r *ConfigURLLimitRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rule, err := r.client.FindURLLimitRule(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing URL limit rule", err.Error())
		return
	}
	if rule == nil {
		resp.Diagnostics.AddError("URL limit rule not found", fmt.Sprintf("No URL limit rule named %q found.", req.ID))
		return
	}
	data := ConfigURLLimitRuleModel{
		Name:    types.StringValue(rule.Name),
		Pattern: types.StringValue(rule.Pattern),
		Limit:   types.Int64Value(rule.Limit),
		Match:   types.Int64Value(rule.Match),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

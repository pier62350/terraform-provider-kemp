// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ datasource.DataSource              = &MatchContentRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &MatchContentRuleDataSource{}
)

func NewMatchContentRuleDataSource() datasource.DataSource { return &MatchContentRuleDataSource{} }

type MatchContentRuleDataSource struct {
	client *loadmaster.Client
}

func (d *MatchContentRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_match_content_rule"
}

func (d *MatchContentRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a system-level content-match rule by name.",
		Attributes: map[string]schema.Attribute{
			"name":            schema.StringAttribute{Required: true, MarkdownDescription: "Rule name."},
			"pattern":         schema.StringAttribute{Computed: true, MarkdownDescription: "Pattern matched against the request."},
			"match_type":      schema.StringAttribute{Computed: true, MarkdownDescription: "Match strategy: `regex`, `prefix`, or `postfix`."},
			"header":          schema.StringAttribute{Computed: true, MarkdownDescription: "Header or scope the pattern is applied to."},
			"include_host":    schema.BoolAttribute{Computed: true},
			"ignore_case":     schema.BoolAttribute{Computed: true},
			"negate":          schema.BoolAttribute{Computed: true},
			"include_query":   schema.BoolAttribute{Computed: true},
			"must_fail":       schema.BoolAttribute{Computed: true},
			"set_on_match":    schema.Int64Attribute{Computed: true},
			"only_on_flag":    schema.Int64Attribute{Computed: true},
			"only_on_no_flag": schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *MatchContentRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*loadmaster.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configure type", fmt.Sprintf("Expected *loadmaster.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *MatchContentRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MatchContentRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := d.client.FindMatchContentRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading match_content_rule", err.Error())
		return
	}
	if rule == nil {
		resp.Diagnostics.AddError("Rule not found", fmt.Sprintf("no MatchContentRule named %q on the LoadMaster", data.Name.ValueString()))
		return
	}

	data.Pattern = types.StringValue(rule.Pattern)
	data.MatchType = types.StringValue(rule.MatchType)
	data.Header = types.StringValue(rule.Header)
	data.IncludeHost = types.BoolValue(rule.AddHost)
	data.IgnoreCase = types.BoolValue(rule.CaseIndependent)
	data.Negate = types.BoolValue(rule.Negate)
	data.IncludeQuery = types.BoolValue(rule.IncludeQuery)
	data.MustFail = types.BoolValue(rule.MustFail)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

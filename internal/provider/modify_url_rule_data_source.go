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
	_ datasource.DataSource              = &ModifyURLRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &ModifyURLRuleDataSource{}
)

func NewModifyURLRuleDataSource() datasource.DataSource { return &ModifyURLRuleDataSource{} }

type ModifyURLRuleDataSource struct {
	client *loadmaster.Client
}

func (d *ModifyURLRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_modify_url_rule"
}

func (d *ModifyURLRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a system-level modify-URL rule by name.",
		Attributes: map[string]schema.Attribute{
			"name":            schema.StringAttribute{Required: true, MarkdownDescription: "Rule name."},
			"pattern":         schema.StringAttribute{Computed: true, MarkdownDescription: "Pattern to match in the URL."},
			"replacement":     schema.StringAttribute{Computed: true, MarkdownDescription: "Replacement string."},
			"only_on_flag":    schema.Int64Attribute{Computed: true},
			"only_on_no_flag": schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *ModifyURLRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ModifyURLRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ModifyURLRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := d.client.FindModifyURLRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading modify_url_rule", err.Error())
		return
	}
	if rule == nil {
		resp.Diagnostics.AddError("Rule not found", fmt.Sprintf("no ModifyURLRule named %q on the LoadMaster", data.Name.ValueString()))
		return
	}

	data.Pattern = types.StringValue(rule.Pattern)
	data.Replacement = types.StringValue(rule.Replacement)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

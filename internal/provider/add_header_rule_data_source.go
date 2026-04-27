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
	_ datasource.DataSource              = &AddHeaderRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &AddHeaderRuleDataSource{}
)

func NewAddHeaderRuleDataSource() datasource.DataSource { return &AddHeaderRuleDataSource{} }

type AddHeaderRuleDataSource struct {
	client *loadmaster.Client
}

func (d *AddHeaderRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_add_header_rule"
}

func (d *AddHeaderRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a system-level add-header rule by name.",
		Attributes: map[string]schema.Attribute{
			"name":            schema.StringAttribute{Required: true, MarkdownDescription: "Rule name."},
			"header":          schema.StringAttribute{Computed: true, MarkdownDescription: "Header field name to add."},
			"replacement":     schema.StringAttribute{Computed: true, MarkdownDescription: "Header value."},
			"only_on_flag":    schema.Int64Attribute{Computed: true},
			"only_on_no_flag": schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *AddHeaderRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AddHeaderRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AddHeaderRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := d.client.FindAddHeaderRule(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading add_header_rule", err.Error())
		return
	}
	if rule == nil {
		resp.Diagnostics.AddError("Rule not found", fmt.Sprintf("no AddHeaderRule named %q on the LoadMaster", data.Name.ValueString()))
		return
	}

	data.Header = types.StringValue(rule.Header)
	data.Replacement = types.StringValue(rule.HeaderValue)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

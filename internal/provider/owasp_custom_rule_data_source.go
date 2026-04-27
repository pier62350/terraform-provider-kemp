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
	_ datasource.DataSource              = &OwaspCustomRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &OwaspCustomRuleDataSource{}
)

func NewOwaspCustomRuleDataSource() datasource.DataSource { return &OwaspCustomRuleDataSource{} }

type OwaspCustomRuleDataSource struct {
	client *loadmaster.Client
}

type OwaspCustomRuleDataSourceModel struct {
	Filename types.String `tfsdk:"filename"`
}

func (d *OwaspCustomRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_owasp_custom_rule"
}

func (d *OwaspCustomRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a custom OWASP/ModSecurity rule file stored on the LoadMaster by filename.",
		Attributes: map[string]schema.Attribute{
			"filename": schema.StringAttribute{Required: true, MarkdownDescription: "Filename of the OWASP custom rule file (e.g. `owaspcust.conf`)."},
		},
	}
}

func (d *OwaspCustomRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OwaspCustomRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OwaspCustomRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify the file exists on the LoadMaster by attempting a download.
	// We discard the content — the data source only confirms existence.
	if err := d.client.VerifyOwaspCustomRule(ctx, data.Filename.ValueString()); err != nil {
		resp.Diagnostics.AddError("OWASP custom rule not found",
			fmt.Sprintf("file %q not found on the LoadMaster: %s", data.Filename.ValueString(), err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ datasource.DataSource              = &OwaspCustomDataDataSource{}
	_ datasource.DataSourceWithConfigure = &OwaspCustomDataDataSource{}
)

func NewOwaspCustomDataDataSource() datasource.DataSource { return &OwaspCustomDataDataSource{} }

type OwaspCustomDataDataSource struct {
	client *loadmaster.Client
}

type OwaspCustomDataDataSourceModel struct {
	Filename types.String `tfsdk:"filename"`
}

func (d *OwaspCustomDataDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_owasp_custom_data"
}

func (d *OwaspCustomDataDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a custom OWASP data file (word/IP list) stored on the LoadMaster by filename.",
		Attributes: map[string]schema.Attribute{
			"filename": schema.StringAttribute{Required: true, MarkdownDescription: "Filename of the OWASP custom data file (e.g. `owasp_cust.data`)."},
		},
	}
}

func (d *OwaspCustomDataDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OwaspCustomDataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OwaspCustomDataDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := d.client.VerifyOwaspCustomData(ctx, data.Filename.ValueString()); err != nil {
		resp.Diagnostics.AddError("OWASP custom data file not found",
			fmt.Sprintf("file %q not found on the LoadMaster: %s", data.Filename.ValueString(), err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

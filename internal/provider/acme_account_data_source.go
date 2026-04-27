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
	_ datasource.DataSource              = &ACMEAccountDataSource{}
	_ datasource.DataSourceWithConfigure = &ACMEAccountDataSource{}
)

func NewACMEAccountDataSource() datasource.DataSource { return &ACMEAccountDataSource{} }

type ACMEAccountDataSource struct {
	client *loadmaster.Client
}

type ACMEAccountDataSourceModel struct {
	ACMEType         types.String `tfsdk:"acme_type"`
	AccountID        types.String `tfsdk:"account_id"`
	AccountDirectory types.String `tfsdk:"account_directory"`
	DirectoryURL     types.String `tfsdk:"directory_url"`
	RenewPeriod      types.Int64  `tfsdk:"renew_period"`
}

func (d *ACMEAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_account"
}

func (d *ACMEAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the ACME account configuration registered on a Kemp LoadMaster for a given provider type.",
		Attributes: map[string]schema.Attribute{
			"acme_type":         schema.StringAttribute{Required: true, MarkdownDescription: "ACME provider: `letsencrypt` or `digicert`."},
			"account_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Registered ACME account identifier."},
			"account_directory": schema.StringAttribute{Computed: true, MarkdownDescription: "Effective ACME directory URL the account is registered against."},
			"directory_url":     schema.StringAttribute{Computed: true, MarkdownDescription: "Configured ACME directory endpoint URL."},
			"renew_period":      schema.Int64Attribute{Computed: true, MarkdownDescription: "Days before expiry at which LoadMaster auto-renews issued certs."},
		},
	}
}

func (d *ACMEAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ACMEAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ACMEAccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiType := acmeTypeToAPI(data.ACMEType.ValueString())

	info, err := d.client.GetACMEAccountInfo(ctx, apiType)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ACME account", err.Error())
		return
	}

	data.AccountID = types.StringValue(info.AccountID)
	data.AccountDirectory = types.StringValue(info.AccountDirectory)

	if url, err := d.client.GetACMEDirectoryURL(ctx, apiType); err == nil {
		data.DirectoryURL = types.StringValue(url)
	} else {
		data.DirectoryURL = types.StringValue("")
	}

	if period, err := d.client.GetACMERenewPeriod(ctx, apiType); err == nil {
		data.RenewPeriod = types.Int64Value(int64(period))
	} else {
		data.RenewPeriod = types.Int64Value(0)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ datasource.DataSource              = &CertificateDataSource{}
	_ datasource.DataSourceWithConfigure = &CertificateDataSource{}
)

func NewCertificateDataSource() datasource.DataSource { return &CertificateDataSource{} }

type CertificateDataSource struct {
	client *loadmaster.Client
}

type CertificateDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

func (d *CertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (d *CertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads metadata for a certificate stored on a Kemp LoadMaster by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the certificate as stored on the LoadMaster."},
			"type": schema.StringAttribute{Computed: true, MarkdownDescription: "Cert type as reported by LoadMaster (`cert`, `pfx`, etc.)."},
		},
	}
}

func (d *CertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := d.client.FindCertificate(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading certificate", err.Error())
		return
	}
	if info == nil {
		resp.Diagnostics.AddError("Certificate not found", fmt.Sprintf("no certificate named %q on the LoadMaster", data.Name.ValueString()))
		return
	}

	data.Name = types.StringValue(info.Name)
	data.Type = types.StringValue(info.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

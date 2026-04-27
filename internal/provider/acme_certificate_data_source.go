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
	_ datasource.DataSource              = &ACMECertificateDataSource{}
	_ datasource.DataSourceWithConfigure = &ACMECertificateDataSource{}
)

func NewACMECertificateDataSource() datasource.DataSource { return &ACMECertificateDataSource{} }

type ACMECertificateDataSource struct {
	client *loadmaster.Client
}

type ACMECertificateDataSourceModel struct {
	Name                  types.String `tfsdk:"name"`
	ACMEType              types.String `tfsdk:"acme_type"`
	DomainName            types.String `tfsdk:"domain_name"`
	ExpiryDate            types.String `tfsdk:"expiry_date"`
	Type                  types.String `tfsdk:"type"`
	SubjectAlternateNames types.String `tfsdk:"subject_alternate_names"`
	HTTPChallengeVS       types.String `tfsdk:"http_challenge_vs"`
}

func (d *ACMECertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_certificate"
}

func (d *ACMECertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an ACME-issued certificate stored on a Kemp LoadMaster by name.",
		Attributes: map[string]schema.Attribute{
			"name":                   schema.StringAttribute{Required: true, MarkdownDescription: "Name of the ACME certificate as stored on the LoadMaster."},
			"acme_type":              schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "ACME provider: `letsencrypt` (default) or `digicert`."},
			"domain_name":            schema.StringAttribute{Computed: true, MarkdownDescription: "Domain name on the issued certificate."},
			"expiry_date":            schema.StringAttribute{Computed: true, MarkdownDescription: "Expiry timestamp as reported by LoadMaster."},
			"type":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Cert algorithm (`rsa`, `ecc`)."},
			"subject_alternate_names": schema.StringAttribute{Computed: true, MarkdownDescription: "Comma-separated SANs as reported by LoadMaster."},
			"http_challenge_vs":      schema.StringAttribute{Computed: true, MarkdownDescription: "VS endpoint (IP:port) used for the HTTP-01 challenge."},
		},
	}
}

func (d *ACMECertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ACMECertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ACMECertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	friendly := data.ACMEType.ValueString()
	if friendly == "" {
		friendly = "letsencrypt"
	}
	friendly = acmeTypeFromAPI(acmeTypeToAPI(friendly))

	info, err := d.client.GetACMECertificate(ctx, data.Name.ValueString(), acmeTypeToAPI(friendly))
	if err != nil {
		resp.Diagnostics.AddError("Error reading ACME certificate", err.Error())
		return
	}

	data.ACMEType = types.StringValue(friendly)
	data.DomainName = types.StringValue(info.DomainName)
	data.ExpiryDate = types.StringValue(info.ExpiryDate)
	data.Type = types.StringValue(info.Type)
	data.SubjectAlternateNames = types.StringValue(info.SubjectAlternateNames)
	data.HTTPChallengeVS = types.StringValue(info.HTTPChallengeVS)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

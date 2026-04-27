// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ datasource.DataSource              = &CertificatesDataSource{}
	_ datasource.DataSourceWithConfigure = &CertificatesDataSource{}
)

func NewCertificatesDataSource() datasource.DataSource { return &CertificatesDataSource{} }

type CertificatesDataSource struct {
	client *loadmaster.Client
}

type CertificatesDataSourceModel struct {
	Certificates types.List `tfsdk:"certificates"`
}

var certObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name": types.StringType,
		"type": types.StringType,
	},
}

func (d *CertificatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (d *CertificatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all certificates stored on the Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"certificates": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "All certificates installed on the LoadMaster.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate name."},
						"type": schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate type as reported by LoadMaster (`cert`, `pfx`, etc.)."},
					},
				},
			},
		},
	}
}

func (d *CertificatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CertificatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certs, err := d.client.ListCertificates(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing certificates", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(certs))
	for _, c := range certs {
		obj, diags := types.ObjectValue(certObjectType.AttrTypes, map[string]attr.Value{
			"name": types.StringValue(c.Name),
			"type": types.StringValue(c.Type),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	listVal, diags := types.ListValue(certObjectType, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Certificates = listVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

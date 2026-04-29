// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var _ datasource.DataSource = &CipherSetDataSource{}

func NewCipherSetDataSource() datasource.DataSource { return &CipherSetDataSource{} }

type CipherSetDataSource struct {
	client *loadmaster.Client
}

type CipherSetDataSourceModel struct {
	Name    types.String `tfsdk:"name"`
	Ciphers types.List   `tfsdk:"ciphers"`
}

func (d *CipherSetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cipher_set"
}

func (d *CipherSetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Reads an existing cipher set from a Kemp LoadMaster by name.

Works for both built-in sets (` + "`Default`" + `, ` + "`Default_NoRc4`" + `, ` + "`BestPractices`" + `, ` + "`Intermediate_compatibility`" + `, ` + "`Backward_compatibility`" + `, ` + "`WUI`" + `, ` + "`FIPS`" + `, ` + "`Legacy`" + `) and custom sets created via ` + "`kemp_cipher_set`" + `. Names are case-sensitive.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Name of the cipher set to look up.",
			},
			"ciphers": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Computed. Ordered list of cipher strings in the set.",
			},
		},
	}
}

func (d *CipherSetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CipherSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CipherSetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cs, err := d.client.GetCipherSet(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading cipher set", err.Error())
		return
	}

	data.Name = types.StringValue(cs.Name)
	var ciphers []string
	if cs.Ciphers != "" {
		for _, c := range strings.Split(cs.Ciphers, ":") {
			c = strings.TrimSpace(c)
			if c != "" {
				ciphers = append(ciphers, c)
			}
		}
	}
	listVal, diags := types.ListValueFrom(ctx, types.StringType, ciphers)
	resp.Diagnostics.Append(diags...)
	data.Ciphers = listVal

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ datasource.DataSource              = &SubVirtualServiceRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &SubVirtualServiceRuleDataSource{}
)

func NewSubVirtualServiceRuleDataSource() datasource.DataSource {
	return &SubVirtualServiceRuleDataSource{}
}

type SubVirtualServiceRuleDataSource struct {
	client *loadmaster.Client
}

func (d *SubVirtualServiceRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sub_virtual_service_rule"
}

func (d *SubVirtualServiceRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Looks up an existing content-switching rule attachment on a SubVS.

Useful for referencing rule attachments that were created outside Terraform (e.g., by the WUI). The SubVS is confirmed to exist; the LoadMaster API does not expose a way to verify that the named rule is actually attached, so this data source trusts the inputs.

Import an existing attachment into managed state with ` + "`kemp_sub_virtual_service_rule`" + `.`,
		Attributes: map[string]schema.Attribute{
			"parent_virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "`Index` of the parent virtual service.",
			},
			"sub_virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "`Index` of the SubVS the rule is attached to.",
			},
			"rule": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the rule that is attached.",
			},
		},
	}
}

func (d *SubVirtualServiceRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubVirtualServiceRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubVirtualServiceRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Confirm the SubVS exists.
	if _, err := d.client.ShowVirtualService(ctx, data.SubVirtualServiceId.ValueString()); err != nil {
		if loadmaster.IsNotFound(err) {
			resp.Diagnostics.AddError("SubVS not found",
				fmt.Sprintf("sub virtual service %q does not exist", data.SubVirtualServiceId.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Error reading SubVS", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

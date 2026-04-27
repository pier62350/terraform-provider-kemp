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
	_ datasource.DataSource              = &VirtualServiceRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &VirtualServiceRuleDataSource{}
)

func NewVirtualServiceRuleDataSource() datasource.DataSource { return &VirtualServiceRuleDataSource{} }

type VirtualServiceRuleDataSource struct {
	client *loadmaster.Client
}

func (d *VirtualServiceRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_service_rule"
}

func (d *VirtualServiceRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks whether a system rule is attached to a virtual service in a given direction. Errors if the attachment does not exist.",
		Attributes: map[string]schema.Attribute{
			"virtual_service_id": schema.StringAttribute{Required: true, MarkdownDescription: "Index of the virtual service."},
			"rule":               schema.StringAttribute{Required: true, MarkdownDescription: "Name of the rule to look up."},
			"direction":          schema.StringAttribute{Required: true, MarkdownDescription: "Direction: `request`, `response`, `responsebody`, or `pre`."},
		},
	}
}

func (d *VirtualServiceRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VirtualServiceRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VirtualServiceRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attached, err := d.client.VSHasRule(ctx, data.VirtualServiceId.ValueString(), data.Rule.ValueString(), data.Direction.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading virtual service rule attachment", err.Error())
		return
	}
	if !attached {
		resp.Diagnostics.AddError("Rule not attached",
			fmt.Sprintf("rule %q is not attached to virtual service %q in direction %q",
				data.Rule.ValueString(), data.VirtualServiceId.ValueString(), data.Direction.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

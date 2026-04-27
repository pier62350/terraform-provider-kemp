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
	_ datasource.DataSource              = &VirtualServiceWafRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &VirtualServiceWafRuleDataSource{}
)

func NewVirtualServiceWafRuleDataSource() datasource.DataSource {
	return &VirtualServiceWafRuleDataSource{}
}

type VirtualServiceWafRuleDataSource struct {
	client *loadmaster.Client
}

func (d *VirtualServiceWafRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_service_waf_rule"
}

func (d *VirtualServiceWafRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a WAF rule attachment on a virtual service (identified by address/port/protocol and rule name).",
		Attributes: map[string]schema.Attribute{
			"virtual_service_address":  schema.StringAttribute{Required: true, MarkdownDescription: "IP address of the virtual service."},
			"virtual_service_port":     schema.StringAttribute{Required: true, MarkdownDescription: "Port of the virtual service."},
			"virtual_service_protocol": schema.StringAttribute{Required: true, MarkdownDescription: "Protocol of the virtual service (`tcp`, `udp`)."},
			"rule":                     schema.StringAttribute{Required: true, MarkdownDescription: "WAF rule path (e.g. `G/ip_reputation`)."},
			"disabled_rules":           schema.StringAttribute{Computed: true, MarkdownDescription: "Comma-separated WAF rule IDs disabled on this VS."},
		},
	}
}

func (d *VirtualServiceWafRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VirtualServiceWafRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VirtualServiceWafRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// LoadMaster does not expose a query for individual WAF rule attachments;
	// we surface the configured state as-is. Use kemp_virtual_service to read
	// the VS's waf_intercept_mode if you need richer WAF status.
	if data.DisabledRules.IsNull() || data.DisabledRules.IsUnknown() {
		data.DisabledRules = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

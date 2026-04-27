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
	_ datasource.DataSource              = &VirtualServiceDataSource{}
	_ datasource.DataSourceWithConfigure = &VirtualServiceDataSource{}
)

func NewVirtualServiceDataSource() datasource.DataSource { return &VirtualServiceDataSource{} }

type VirtualServiceDataSource struct {
	client *loadmaster.Client
}

func (d *VirtualServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_service"
}

func (d *VirtualServiceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Kemp LoadMaster virtual service by its `Index`.",
		Attributes: map[string]schema.Attribute{
			"id":                        schema.StringAttribute{Required: true, MarkdownDescription: "LoadMaster `Index` of the virtual service."},
			"address":                   schema.StringAttribute{Computed: true},
			"port":                      schema.StringAttribute{Computed: true},
			"protocol":                  schema.StringAttribute{Computed: true},
			"type":                      schema.StringAttribute{Computed: true},
			"nickname":                  schema.StringAttribute{Computed: true},
			"enabled":                   schema.BoolAttribute{Computed: true},
			"ssl_acceleration":          schema.BoolAttribute{Computed: true},
			"cert_files":                schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"schedule":                  schema.StringAttribute{Computed: true},
			"persist_timeout":           schema.StringAttribute{Computed: true},
			"idletime":                  schema.Int64Attribute{Computed: true},
			"force_l7":                  schema.BoolAttribute{Computed: true},
			"check_type":                schema.StringAttribute{Computed: true},
			"check_port":                schema.StringAttribute{Computed: true},
			"chk_interval":              schema.Int64Attribute{Computed: true},
			"chk_timeout":               schema.Int64Attribute{Computed: true},
			"chk_retry_count":           schema.Int64Attribute{Computed: true},
			"bandwidth":                 schema.Int64Attribute{Computed: true},
			"conns_per_sec_limit":       schema.Int64Attribute{Computed: true},
			"requests_per_sec_limit":    schema.Int64Attribute{Computed: true},
			"max_conns_limit":           schema.Int64Attribute{Computed: true},
			"esp_enabled":               schema.BoolAttribute{Computed: true},
			"esp_allowed_hosts":         schema.StringAttribute{Computed: true},
			"esp_allowed_directories":   schema.StringAttribute{Computed: true},
			"esp_input_auth_mode":       schema.StringAttribute{Computed: true},
			"esp_output_auth_mode":      schema.StringAttribute{Computed: true},
			"esp_include_nested_groups": schema.BoolAttribute{Computed: true},
			"esp_display_pub_priv":      schema.BoolAttribute{Computed: true},
			"esp_logs":                  schema.BoolAttribute{Computed: true},
			"waf_intercept_mode":        schema.StringAttribute{Computed: true},
			"waf_blocking_paranoia":     schema.Int64Attribute{Computed: true},
			"waf_alert_threshold":       schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *VirtualServiceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*loadmaster.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected *loadmaster.Client, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *VirtualServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VirtualServiceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := d.client.ShowVirtualService(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading virtual service", err.Error())
		return
	}

	r := &VirtualServiceResource{}
	resp.Diagnostics.Append(r.writeState(ctx, vs, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

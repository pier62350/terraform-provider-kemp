// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ datasource.DataSource              = &RealServerDataSource{}
	_ datasource.DataSourceWithConfigure = &RealServerDataSource{}
)

func NewRealServerDataSource() datasource.DataSource { return &RealServerDataSource{} }

type RealServerDataSource struct {
	client *loadmaster.Client
}

func (d *RealServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_real_server"
}

func (d *RealServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a real server attached to a Kemp virtual service.",
		Attributes: map[string]schema.Attribute{
			"id":                 schema.Int32Attribute{Required: true, MarkdownDescription: "LoadMaster `RsIndex`."},
			"virtual_service_id": schema.StringAttribute{Required: true, MarkdownDescription: "Index of the parent virtual service."},
			"address":            schema.StringAttribute{Computed: true},
			"port":               schema.StringAttribute{Computed: true},
			"weight":             schema.Int32Attribute{Computed: true},
			"forward":            schema.StringAttribute{Computed: true},
			"enable":             schema.BoolAttribute{Computed: true},
			"limit":              schema.Int32Attribute{Computed: true},
			"critical":           schema.BoolAttribute{Computed: true},
			"follow":             schema.Int32Attribute{Computed: true},
			"dns_name":           schema.StringAttribute{Computed: true},
		},
	}
}

func (d *RealServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RealServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RealServerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rs, err := d.client.ShowRealServer(ctx, data.VirtualServiceId.ValueString(), strconv.Itoa(int(data.Id.ValueInt32())))
	if err != nil {
		resp.Diagnostics.AddError("Error reading real server", err.Error())
		return
	}

	data.Id = types.Int32Value(rs.RsIndex)
	data.VirtualServiceId = types.StringValue(strconv.Itoa(int(rs.VSIndex)))
	data.Address = types.StringValue(rs.Address)
	data.Port = types.StringValue(strconv.Itoa(int(rs.Port)))
	data.Weight = types.Int32Value(rs.Weight)
	data.Forward = types.StringValue(rs.Forward)
	data.Limit = types.Int32Value(rs.Limit)
	data.Follow = types.Int32Value(rs.Follow)
	data.DnsName = types.StringValue(rs.DnsName)
	if rs.Enable != nil {
		data.Enable = types.BoolValue(*rs.Enable)
	} else {
		data.Enable = types.BoolValue(false)
	}
	if rs.Critical != nil {
		data.Critical = types.BoolValue(*rs.Critical)
	} else {
		data.Critical = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

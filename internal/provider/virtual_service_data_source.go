// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
			"id":               schema.StringAttribute{Required: true, MarkdownDescription: "LoadMaster `Index` of the virtual service."},
			"address":          schema.StringAttribute{Computed: true},
			"port":             schema.StringAttribute{Computed: true},
			"protocol":         schema.StringAttribute{Computed: true},
			"type":             schema.StringAttribute{Computed: true},
			"nickname":         schema.StringAttribute{Computed: true},
			"enabled":          schema.BoolAttribute{Computed: true},
			"ssl_acceleration": schema.BoolAttribute{Computed: true},
			"cert_files":       schema.ListAttribute{Computed: true, ElementType: types.StringType},
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

	data.Id = types.StringValue(strconv.Itoa(int(vs.Index)))
	data.Address = types.StringValue(vs.Address)
	data.Port = types.StringValue(vs.Port)
	data.Protocol = types.StringValue(vs.Protocol)
	data.Type = types.StringValue(vs.VSType)
	data.Nickname = types.StringValue(vs.NickName)
	if vs.Enable != nil {
		data.Enabled = types.BoolValue(*vs.Enable)
	} else {
		data.Enabled = types.BoolValue(false)
	}
	if vs.SSLAcceleration != nil {
		data.SSLAcceleration = types.BoolValue(*vs.SSLAcceleration)
	} else {
		data.SSLAcceleration = types.BoolValue(false)
	}
	var certs []string
	if vs.CertFile != "" {
		certs = strings.Split(vs.CertFile, ",")
		for i := range certs {
			certs[i] = strings.TrimSpace(certs[i])
		}
	}
	listVal, listDiags := types.ListValueFrom(ctx, types.StringType, certs)
	resp.Diagnostics.Append(listDiags...)
	data.CertFiles = listVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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

var _ datasource.DataSource = &InterfaceDataSource{}

func NewInterfaceDataSource() datasource.DataSource { return &InterfaceDataSource{} }

type InterfaceDataSource struct {
	client *loadmaster.Client
}

type InterfaceDataSourceModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	IPAddress           types.String `tfsdk:"ip_address"`
	Mtu                 types.Int64  `tfsdk:"mtu"`
	AdditionalAddresses types.List   `tfsdk:"additional_addresses"`
	InterfaceType       types.String `tfsdk:"interface_type"`
	GeoTrafficEnabled   types.Bool   `tfsdk:"geo_traffic_enabled"`
	DefaultInterface    types.Bool   `tfsdk:"default_interface"`
}

func (d *InterfaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface"
}

func (d *InterfaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a network interface from a Kemp LoadMaster by its zero-based index.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** Zero-based index of the interface (e.g. `0` for eth0).",
			},
			"ip_address": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Primary IP address of the interface in CIDR notation.",
			},
			"mtu": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Computed. MTU of the interface.",
			},
			"additional_addresses": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Computed. Additional IP addresses assigned to the interface.",
			},
			"interface_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Interface type (e.g. `Port`).",
			},
			"geo_traffic_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Whether GEO traffic is enabled on this interface.",
			},
			"default_interface": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Whether this is the default interface.",
			},
		},
	}
}

func (d *InterfaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InterfaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idStr := strconv.FormatInt(data.ID.ValueInt64(), 10)
	iface, err := d.client.ShowInterface(ctx, idStr)
	if err != nil {
		resp.Diagnostics.AddError("Error reading interface", err.Error())
		return
	}

	data.IPAddress = types.StringValue(iface.IPAddress)
	data.InterfaceType = types.StringValue(iface.InterfaceType)
	data.GeoTrafficEnabled = types.BoolValue(iface.GeoTrafficEnable)
	data.DefaultInterface = types.BoolValue(iface.DefaultInterface)

	if mtu, err := strconv.ParseInt(iface.Mtu, 10, 64); err == nil {
		data.Mtu = types.Int64Value(mtu)
	} else {
		data.Mtu = types.Int64Value(0)
	}

	addrs := iface.AdditionalAddresses
	if addrs == nil {
		addrs = []string{}
	}
	addrList, diags := types.ListValueFrom(ctx, types.StringType, addrs)
	resp.Diagnostics.Append(diags...)
	data.AdditionalAddresses = addrList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

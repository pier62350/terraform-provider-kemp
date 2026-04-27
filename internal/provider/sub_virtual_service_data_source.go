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
	_ datasource.DataSource              = &SubVirtualServiceDataSource{}
	_ datasource.DataSourceWithConfigure = &SubVirtualServiceDataSource{}
)

func NewSubVirtualServiceDataSource() datasource.DataSource { return &SubVirtualServiceDataSource{} }

type SubVirtualServiceDataSource struct {
	client *loadmaster.Client
}

type SubVirtualServiceDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	ParentId        types.String `tfsdk:"parent_id"`
	Nickname        types.String `tfsdk:"nickname"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Type            types.String `tfsdk:"type"`
	SSLAcceleration types.Bool   `tfsdk:"ssl_acceleration"`
	CertFiles       types.List   `tfsdk:"cert_files"`

	EspEnabled             types.Bool   `tfsdk:"esp_enabled"`
	EspAllowedHosts        types.String `tfsdk:"esp_allowed_hosts"`
	EspAllowedDirectories  types.String `tfsdk:"esp_allowed_directories"`
	EspInputAuthMode       types.String `tfsdk:"esp_input_auth_mode"`
	EspOutputAuthMode      types.String `tfsdk:"esp_output_auth_mode"`
	EspIncludeNestedGroups types.Bool   `tfsdk:"esp_include_nested_groups"`
	EspDisplayPubPriv      types.Bool   `tfsdk:"esp_display_pub_priv"`
	EspLogs                types.Bool   `tfsdk:"esp_logs"`

	WafInterceptMode    types.String `tfsdk:"waf_intercept_mode"`
	WafBlockingParanoia types.Int64  `tfsdk:"waf_blocking_paranoia"`
	WafAlertThreshold   types.Int64  `tfsdk:"waf_alert_threshold"`
}

func (d *SubVirtualServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sub_virtual_service"
}

func (d *SubVirtualServiceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Sub-Virtual Service by its LoadMaster `Index`.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Required: true, MarkdownDescription: "LoadMaster `Index` of the SubVS."},
			"parent_id":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "`Index` of the parent virtual service."},
			"nickname":         schema.StringAttribute{Computed: true, MarkdownDescription: "Friendly name shown in the WUI."},
			"enabled":          schema.BoolAttribute{Computed: true},
			"type":             schema.StringAttribute{Computed: true, MarkdownDescription: "VS type (`gen`, `http`, etc.)."},
			"ssl_acceleration": schema.BoolAttribute{Computed: true},
			"cert_files":       schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Certificates attached to this SubVS."},

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

func (d *SubVirtualServiceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubVirtualServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubVirtualServiceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := d.client.ShowVirtualService(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading sub-virtual service", err.Error())
		return
	}

	data.Id = types.StringValue(strconv.Itoa(int(vs.Index)))
	data.Nickname = types.StringValue(vs.NickName)
	data.Type = types.StringValue(vs.VSType)
	if vs.Enable != nil {
		data.Enabled = types.BoolValue(*vs.Enable)
	} else {
		data.Enabled = types.BoolValue(false)
	}
	data.SSLAcceleration = boolFromPtr(vs.SSLAcceleration)

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

	data.EspEnabled = boolFromPtr(vs.EspEnabled)
	data.EspAllowedHosts = types.StringValue(vs.AllowedHosts)
	data.EspAllowedDirectories = types.StringValue(vs.AllowedDirectories)
	data.EspInputAuthMode = types.StringValue(vs.InputAuthMode)
	data.EspOutputAuthMode = types.StringValue(vs.OutputAuthMode)
	data.EspIncludeNestedGroups = boolFromPtr(vs.IncludeNestedGroups)
	data.EspDisplayPubPriv = boolFromPtr(vs.DisplayPubPriv)
	data.EspLogs = boolFromPtr(vs.EspLogs)
	data.WafInterceptMode = types.StringValue(vs.InterceptMode)
	data.WafBlockingParanoia = int64FromPtr(vs.BlockingParanoia)
	data.WafAlertThreshold = int64FromPtr(vs.AlertThreshold)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

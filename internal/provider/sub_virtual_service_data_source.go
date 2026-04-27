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

	Schedule            types.String `tfsdk:"schedule"`
	PersistTimeout      types.String `tfsdk:"persist_timeout"`
	Idletime            types.Int64  `tfsdk:"idletime"`
	ForceL7             types.Bool   `tfsdk:"force_l7"`
	Bandwidth           types.Int64  `tfsdk:"bandwidth"`
	ConnsPerSecLimit    types.Int64  `tfsdk:"conns_per_sec_limit"`
	RequestsPerSecLimit types.Int64  `tfsdk:"requests_per_sec_limit"`
	MaxConnsLimit       types.Int64  `tfsdk:"max_conns_limit"`

	CheckType            types.String `tfsdk:"check_type"`
	CheckPort            types.String `tfsdk:"check_port"`
	ChkInterval          types.Int64  `tfsdk:"chk_interval"`
	ChkTimeout           types.Int64  `tfsdk:"chk_timeout"`
	ChkRetryCount        types.Int64  `tfsdk:"chk_retry_count"`
	NeedHostName         types.Bool   `tfsdk:"need_host_name"`
	CheckUseHTTP11       types.Bool   `tfsdk:"check_use_http11"`
	CheckUseGet          types.String `tfsdk:"check_use_get"`
	MatchLen             types.Int64  `tfsdk:"match_len"`
	EnhancedHealthChecks types.Bool   `tfsdk:"enhanced_health_checks"`

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
			"nickname":         schema.StringAttribute{Computed: true},
			"enabled":          schema.BoolAttribute{Computed: true},
			"type":             schema.StringAttribute{Computed: true},
			"ssl_acceleration": schema.BoolAttribute{Computed: true},
			"cert_files":       schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"schedule":                schema.StringAttribute{Computed: true},
			"persist_timeout":         schema.StringAttribute{Computed: true},
			"idletime":                schema.Int64Attribute{Computed: true},
			"force_l7":                schema.BoolAttribute{Computed: true},
			"bandwidth":               schema.Int64Attribute{Computed: true},
			"conns_per_sec_limit":     schema.Int64Attribute{Computed: true},
			"requests_per_sec_limit":  schema.Int64Attribute{Computed: true},
			"max_conns_limit":         schema.Int64Attribute{Computed: true},
			"check_type":              schema.StringAttribute{Computed: true},
			"check_port":              schema.StringAttribute{Computed: true},
			"chk_interval":            schema.Int64Attribute{Computed: true},
			"chk_timeout":             schema.Int64Attribute{Computed: true},
			"chk_retry_count":         schema.Int64Attribute{Computed: true},
			"need_host_name":          schema.BoolAttribute{Computed: true},
			"check_use_http11":        schema.BoolAttribute{Computed: true},
			"check_use_get":           schema.StringAttribute{Computed: true},
			"match_len":               schema.Int64Attribute{Computed: true},
			"enhanced_health_checks":  schema.BoolAttribute{Computed: true},
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

	data.Schedule = types.StringValue(vs.Schedule)
	data.PersistTimeout = types.StringValue(vs.PersistTimeout)
	data.Idletime = int64FromPtr(vs.Idletime)
	data.ForceL7 = boolFromPtr(vs.ForceL7)
	data.Bandwidth = int64FromPtr(vs.Bandwidth)
	data.ConnsPerSecLimit = int64FromPtr(vs.ConnsPerSecLimit)
	data.RequestsPerSecLimit = int64FromPtr(vs.RequestsPerSecLimit)
	data.MaxConnsLimit = int64FromPtr(vs.MaxConnsLimit)

	data.CheckType = types.StringValue(vs.CheckType)
	data.CheckPort = types.StringValue(vs.CheckPort)
	data.ChkInterval = int64FromPtr(vs.ChkInterval)
	data.ChkTimeout = int64FromPtr(vs.ChkTimeout)
	data.ChkRetryCount = int64FromPtr(vs.ChkRetryCount)
	data.NeedHostName = boolFromPtr(vs.NeedHostName)
	data.CheckUseHTTP11 = boolFromPtr(vs.CheckUseHTTP11)
	data.CheckUseGet = types.StringValue(checkUseGetFromAPI(vs.CheckUseGet))
	data.MatchLen = int64FromPtr(vs.MatchLen)
	data.EnhancedHealthChecks = boolFromPtr(vs.EnhancedHealthChecks)

	data.EspEnabled = boolFromPtr(vs.EspEnabled)
	data.EspAllowedHosts = types.StringValue(vs.AllowedHosts)
	data.EspAllowedDirectories = types.StringValue(vs.AllowedDirectories)
	data.EspInputAuthMode = types.StringValue(espInputAuthModeFromAPI(vs.InputAuthMode))
	data.EspOutputAuthMode = types.StringValue(espOutputAuthModeFromAPI(vs.OutputAuthMode))
	data.EspIncludeNestedGroups = boolFromPtr(vs.IncludeNestedGroups)
	data.EspDisplayPubPriv = boolFromPtr(vs.DisplayPubPriv)
	data.EspLogs = boolFromPtr(vs.EspLogs)
	data.WafInterceptMode = types.StringValue(wafInterceptModeFromAPI(vs.InterceptMode))
	data.WafBlockingParanoia = int64FromPtr(vs.BlockingParanoia)
	data.WafAlertThreshold = int64FromPtr(vs.AlertThreshold)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

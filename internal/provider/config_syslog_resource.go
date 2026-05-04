// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigSyslogResource{}
	_ resource.ResourceWithImportState = &ConfigSyslogResource{}
)

func NewConfigSyslogResource() resource.Resource { return &ConfigSyslogResource{} }

type ConfigSyslogResource struct {
	client *loadmaster.Client
}

type ConfigSyslogResourceModel struct {
	Level types.String `tfsdk:"level"`
	Hosts types.List   `tfsdk:"hosts"`
}

func (r *ConfigSyslogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_syslog"
}

func (r *ConfigSyslogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the syslog destination hosts for a given severity level on a Kemp LoadMaster.

Each severity level is managed independently. Destroying this resource clears all hosts for that level.

Valid levels: ` + "`notice`" + `, ` + "`warning`" + `, ` + "`error`" + `, ` + "`critical`" + `, ` + "`alert`" + `, ` + "`emergency`" + `.`,
		Attributes: map[string]schema.Attribute{
			"level": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Syslog severity level. One of: `notice`, `warning`, `error`, `critical`, `alert`, `emergency`. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"hosts": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "**Required.** List of syslog destination hosts in `host:port` format (e.g. `10.0.0.5:514`).",
			},
		},
	}
}

func (r *ConfigSyslogResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*loadmaster.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configure type", fmt.Sprintf("Expected *loadmaster.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func hostsFromModel(ctx context.Context, data ConfigSyslogResourceModel) []string {
	var raw []string
	data.Hosts.ElementsAs(ctx, &raw, false)
	return raw
}

func (r *ConfigSyslogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigSyslogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hosts := hostsFromModel(ctx, data)
	if err := r.client.SetSyslogHosts(ctx, data.Level.ValueString(), hosts); err != nil {
		resp.Diagnostics.AddError("Error configuring syslog", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigSyslogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigSyslogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hosts, err := r.client.GetSyslogHosts(ctx, data.Level.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading syslog", err.Error())
		return
	}

	if len(hosts) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	hostVals := make([]types.String, len(hosts))
	for i, h := range hosts {
		hostVals[i] = types.StringValue(strings.TrimSpace(h))
	}
	hostList, diags := types.ListValueFrom(ctx, types.StringType, hostVals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Hosts = hostList
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigSyslogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigSyslogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hosts := hostsFromModel(ctx, data)
	if err := r.client.SetSyslogHosts(ctx, data.Level.ValueString(), hosts); err != nil {
		resp.Diagnostics.AddError("Error updating syslog", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigSyslogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigSyslogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetSyslogHosts(ctx, data.Level.ValueString(), []string{}); err != nil {
		resp.Diagnostics.AddError("Error clearing syslog hosts", err.Error())
	}
}

func (r *ConfigSyslogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	level := req.ID
	hosts, err := r.client.GetSyslogHosts(ctx, level)
	if err != nil {
		resp.Diagnostics.AddError("Error importing syslog", err.Error())
		return
	}

	hostVals := make([]types.String, len(hosts))
	for i, h := range hosts {
		hostVals[i] = types.StringValue(strings.TrimSpace(h))
	}
	hostList, diags := types.ListValueFrom(ctx, types.StringType, hostVals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := ConfigSyslogResourceModel{
		Level: types.StringValue(level),
		Hosts: hostList,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

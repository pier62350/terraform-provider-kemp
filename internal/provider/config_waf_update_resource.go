// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigWAFUpdateResource{}
	_ resource.ResourceWithImportState = &ConfigWAFUpdateResource{}
)

func NewConfigWAFUpdateResource() resource.Resource {
	return &ConfigWAFUpdateResource{}
}

type ConfigWAFUpdateResource struct {
	client *loadmaster.Client
}

type ConfigWAFUpdateModel struct {
	ID          types.String `tfsdk:"id"`
	AutoUpdate  types.Bool   `tfsdk:"auto_update"`
	AutoInstall types.Bool   `tfsdk:"auto_install"`
	InstallHour types.Int64  `tfsdk:"install_hour"`
}

func (r *ConfigWAFUpdateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_waf_update"
}

func (r *ConfigWAFUpdateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the commercial WAF rule auto-update schedule on a Kemp LoadMaster.

This is a singleton resource. Import with: ` + "`terraform import kemp_config_waf_update.main loadmaster`",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `loadmaster`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"auto_update": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Whether WAF rules should be automatically downloaded when updates are available.",
			},
			"auto_install": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Whether downloaded WAF rule updates should be automatically installed.",
			},
			"install_hour": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Hour of the day (0–23) at which WAF rule updates are installed. Defaults to `0`.",
			},
		},
	}
}

func (r *ConfigWAFUpdateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigWAFUpdateResource) applyAll(ctx context.Context, data ConfigWAFUpdateModel) error {
	if err := r.client.SetWAFAutoUpdate(ctx, data.AutoUpdate.ValueBool()); err != nil {
		return fmt.Errorf("setwafautoupdate: %w", err)
	}
	if err := r.client.SetWAFAutoInstall(ctx, data.AutoInstall.ValueBool()); err != nil {
		return fmt.Errorf("setwafautoinstall: %w", err)
	}
	if err := r.client.SetWAFInstallHour(ctx, data.InstallHour.ValueInt64()); err != nil {
		return fmt.Errorf("setwafinstalltime: %w", err)
	}
	return nil
}

func (r *ConfigWAFUpdateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigWAFUpdateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyAll(ctx, data); err != nil {
		resp.Diagnostics.AddError("Error configuring WAF update settings", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWAFUpdateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigWAFUpdateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetWAFSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading WAF settings", err.Error())
		return
	}

	data.ID = types.StringValue("loadmaster")

	// getwafsettings may not surface all auto-update fields reliably.
	// Only update state fields when the API returns a value; otherwise keep
	// the current state to avoid spurious drift.
	if settings.AutoUpdate != nil {
		data.AutoUpdate = types.BoolValue(*settings.AutoUpdate)
	}
	if settings.AutoInstall != nil {
		data.AutoInstall = types.BoolValue(*settings.AutoInstall)
	}
	if settings.InstallHour != nil {
		data.InstallHour = types.Int64Value(int64(*settings.InstallHour))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWAFUpdateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigWAFUpdateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyAll(ctx, data); err != nil {
		resp.Diagnostics.AddError("Error updating WAF update settings", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWAFUpdateResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton — removing from state is sufficient; LoadMaster config is unchanged.
}

func (r *ConfigWAFUpdateResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settings, err := r.client.GetWAFSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing WAF update settings", err.Error())
		return
	}

	data := ConfigWAFUpdateModel{
		ID:          types.StringValue("loadmaster"),
		AutoUpdate:  types.BoolValue(false),
		AutoInstall: types.BoolValue(false),
		InstallHour: types.Int64Value(0),
	}
	if settings.AutoUpdate != nil {
		data.AutoUpdate = types.BoolValue(*settings.AutoUpdate)
	}
	if settings.AutoInstall != nil {
		data.AutoInstall = types.BoolValue(*settings.AutoInstall)
	}
	if settings.InstallHour != nil {
		data.InstallHour = types.Int64Value(int64(*settings.InstallHour))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

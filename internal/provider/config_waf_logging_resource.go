// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigWAFLoggingResource{}
	_ resource.ResourceWithImportState = &ConfigWAFLoggingResource{}
)

func NewConfigWAFLoggingResource() resource.Resource {
	return &ConfigWAFLoggingResource{}
}

type ConfigWAFLoggingResource struct {
	client *loadmaster.Client
}

type ConfigWAFLoggingModel struct {
	ID        types.String `tfsdk:"id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	RemoteURI types.String `tfsdk:"remote_uri"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	LogFormat types.String `tfsdk:"log_format"`
}

func (r *ConfigWAFLoggingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_waf_logging"
}

func (r *ConfigWAFLoggingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages WAF remote logging settings on a Kemp LoadMaster.

This is a singleton resource. Import with: ` + "`terraform import kemp_config_waf_logging.main loadmaster`",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `loadmaster`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Whether WAF remote logging is enabled.",
			},
			"remote_uri": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "URI of the remote log receiver (e.g. `https://siem.example.com/waf`). Required when `enabled` is `true`.",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Username for authenticating to the remote log receiver.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password for authenticating to the remote log receiver. Write-only — not read back from the API.",
			},
			"log_format": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "WAF log format. Common values: `cef`, `leef`, `w3c`.",
			},
		},
	}
}

func (r *ConfigWAFLoggingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigWAFLoggingResource) applyLogging(ctx context.Context, data ConfigWAFLoggingModel) error {
	if data.Enabled.ValueBool() {
		if err := r.client.EnableWAFRemoteLogging(
			ctx,
			data.RemoteURI.ValueString(),
			data.Username.ValueString(),
			data.Password.ValueString(),
		); err != nil {
			return fmt.Errorf("enablewafremotelogging: %w", err)
		}
	} else {
		if err := r.client.DisableWAFRemoteLogging(ctx); err != nil {
			return fmt.Errorf("disablewafremotelogging: %w", err)
		}
	}

	if !data.LogFormat.IsNull() && !data.LogFormat.IsUnknown() && data.LogFormat.ValueString() != "" {
		if err := r.client.SetWAFLogFormat(ctx, data.LogFormat.ValueString()); err != nil {
			return fmt.Errorf("setwaflogformat: %w", err)
		}
	}
	return nil
}

func (r *ConfigWAFLoggingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigWAFLoggingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyLogging(ctx, data); err != nil {
		resp.Diagnostics.AddError("Error configuring WAF logging", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWAFLoggingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigWAFLoggingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetWAFSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading WAF logging settings", err.Error())
		return
	}

	data.ID = types.StringValue("loadmaster")

	if settings.RemoteLoggingEnabled != nil {
		data.Enabled = types.BoolValue(*settings.RemoteLoggingEnabled)
	}
	if settings.RemoteURI != "" {
		data.RemoteURI = types.StringValue(settings.RemoteURI)
	}
	if settings.LogFormat != "" {
		data.LogFormat = types.StringValue(settings.LogFormat)
	}

	// password is write-only — preserve whatever is already in state; do not
	// read back from the API (the API never returns it).

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWAFLoggingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigWAFLoggingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyLogging(ctx, data); err != nil {
		resp.Diagnostics.AddError("Error updating WAF logging settings", err.Error())
		return
	}
	data.ID = types.StringValue("loadmaster")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWAFLoggingResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton — removing from state is sufficient; LoadMaster config is unchanged.
}

func (r *ConfigWAFLoggingResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settings, err := r.client.GetWAFSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing WAF logging settings", err.Error())
		return
	}

	data := ConfigWAFLoggingModel{
		ID:        types.StringValue("loadmaster"),
		Enabled:   types.BoolValue(false),
		RemoteURI: types.StringNull(),
		Username:  types.StringNull(),
		// password is write-only — cannot be recovered on import; must be set
		// after import via configuration.
		Password:  types.StringNull(),
		LogFormat: types.StringNull(),
	}

	if settings.RemoteLoggingEnabled != nil {
		data.Enabled = types.BoolValue(*settings.RemoteLoggingEnabled)
	}
	if settings.RemoteURI != "" {
		data.RemoteURI = types.StringValue(settings.RemoteURI)
	}
	if settings.LogFormat != "" {
		data.LogFormat = types.StringValue(settings.LogFormat)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

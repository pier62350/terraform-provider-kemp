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
	_ resource.Resource                = &ConfigHAResource{}
	_ resource.ResourceWithImportState = &ConfigHAResource{}
)

func NewConfigHAResource() resource.Resource { return &ConfigHAResource{} }

type ConfigHAResource struct {
	client *loadmaster.Client
}

type ConfigHAResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Mode           types.Int64  `tfsdk:"mode"`
	PartnerAddress types.String `tfsdk:"partner_address"`
	SharedAddress  types.String `tfsdk:"shared_address"`
	Secret         types.String `tfsdk:"secret"`
}

func (r *ConfigHAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_ha"
}

func (r *ConfigHAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the High Availability (HA) pair configuration on a Kemp LoadMaster.

This is a singleton resource. Import with: ` + "`terraform import kemp_config_ha.main loadmaster`" + `

**Note:** The ` + "`secret`" + ` attribute is write-only. It is not read back from the LoadMaster and will be ` + "`null`" + ` after import.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `loadmaster`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"mode": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "**Required.** HA mode. `0` = non-HA, `1` = primary (HA1), `2` = secondary (HA2).",
			},
			"partner_address": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional. Partner LoadMaster IP address.",
			},
			"shared_address": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional. Shared virtual IP address for the HA pair.",
			},
			"secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Optional. Partner communication shared secret. This value is write-only and is never read back from the LoadMaster.",
			},
		},
	}
}

func (r *ConfigHAResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigHAResource) applyConfig(ctx context.Context, data ConfigHAResourceModel) error {
	if err := r.client.SetHAMode(ctx, data.Mode.ValueInt64()); err != nil {
		return fmt.Errorf("set HA mode: %w", err)
	}
	if !data.PartnerAddress.IsNull() && !data.PartnerAddress.IsUnknown() {
		if err := r.client.SetHAPartnerAddress(ctx, data.PartnerAddress.ValueString()); err != nil {
			return fmt.Errorf("set partner address: %w", err)
		}
	}
	if !data.SharedAddress.IsNull() && !data.SharedAddress.IsUnknown() {
		if err := r.client.SetHASharedAddress(ctx, data.SharedAddress.ValueString()); err != nil {
			return fmt.Errorf("set shared address: %w", err)
		}
	}
	if !data.Secret.IsNull() && !data.Secret.IsUnknown() && data.Secret.ValueString() != "" {
		if err := r.client.SetHASecret(ctx, data.Secret.ValueString()); err != nil {
			return fmt.Errorf("set HA secret: %w", err)
		}
	}
	return nil
}

func writeHAState(ha *loadmaster.HAConfig, data *ConfigHAResourceModel) {
	data.ID = types.StringValue("loadmaster")
	data.Mode = types.Int64Value(ha.Mode)
	if ha.PartnerAddress != "" {
		data.PartnerAddress = types.StringValue(ha.PartnerAddress)
	}
	if ha.SharedAddress != "" {
		data.SharedAddress = types.StringValue(ha.SharedAddress)
	}
	// Secret is write-only — never populate from read.
}

func (r *ConfigHAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigHAResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Error configuring HA", err.Error())
		return
	}

	ha, err := r.client.GetHAConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HA config after create", err.Error())
		return
	}
	writeHAState(ha, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigHAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigHAResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ha, err := r.client.GetHAConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HA config", err.Error())
		return
	}
	// Preserve secret from state (write-only).
	secret := data.Secret
	writeHAState(ha, &data)
	data.Secret = secret
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigHAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigHAResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Error updating HA config", err.Error())
		return
	}

	ha, err := r.client.GetHAConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HA config after update", err.Error())
		return
	}
	secret := data.Secret
	writeHAState(ha, &data)
	data.Secret = secret
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigHAResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton — removing from state is sufficient; LoadMaster config is unchanged.
}

func (r *ConfigHAResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ha, err := r.client.GetHAConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing HA config", err.Error())
		return
	}
	var data ConfigHAResourceModel
	writeHAState(ha, &data)
	// Secret is not imported — leave as null.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

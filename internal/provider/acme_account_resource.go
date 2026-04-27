// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ACMEAccountResource{}
	_ resource.ResourceWithImportState = &ACMEAccountResource{}
)

func NewACMEAccountResource() resource.Resource { return &ACMEAccountResource{} }

type ACMEAccountResource struct {
	client *loadmaster.Client
}

type ACMEAccountModel struct {
	ACMEType         types.String `tfsdk:"acme_type"`
	Email            types.String `tfsdk:"email"`
	DirectoryURL     types.String `tfsdk:"directory_url"`
	RenewPeriod      types.Int64  `tfsdk:"renew_period"`
	KID              types.String `tfsdk:"kid"`
	HMACKey          types.String `tfsdk:"hmac_key"`
	AccountID        types.String `tfsdk:"account_id"`
	AccountDirectory types.String `tfsdk:"account_directory"`
}

func (r *ACMEAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_account"
}

func (r *ACMEAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Configures and registers an ACME account on the LoadMaster (Let's Encrypt or DigiCert). This is a prerequisite for ` + "`kemp_acme_certificate`" + `.

The Create flow runs three commands in order: ` + "`setacmedirectoryurl`" + ` (if specified), ` + "`registeracmeaccount`" + `, then ` + "`setacmerenewperiod`" + ` (if specified). Update only re-runs the directory-URL and renew-period sets.

Delete calls ` + "`delacmeconfig`" + ` which only succeeds if there are no ACME certs of this type still on the LoadMaster.`,
		Attributes: map[string]schema.Attribute{
			"acme_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ACME provider: `1` for Let's Encrypt, `2` for DigiCert.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Account registration email (Let's Encrypt). Cannot be changed after registration without delete+recreate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"directory_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ACME directory endpoint URL. Set to `https://acme-staging-v02.api.letsencrypt.org/directory` for Let's Encrypt staging, omit for the LoadMaster's default.",
				Default:             stringdefault.StaticString(""),
			},
			"renew_period": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Days before expiry at which LoadMaster auto-renews issued certs (1-60).",
			},
			"kid": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "DigiCert Key ID (`setacmekid`). Only valid when `acme_type = \"2\"`. Write-only — LoadMaster does not return this value on read.",
			},
			"hmac_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "DigiCert HMAC key (`setacmehmac`). Only valid when `acme_type = \"2\"`. Write-only — LoadMaster does not return this value on read.",
			},
			"account_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Registered ACME account identifier."},
			"account_directory": schema.StringAttribute{Computed: true, MarkdownDescription: "Effective ACME directory URL the account is registered against."},
		},
	}
}

func (r *ACMEAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ACMEAccountResource) refresh(ctx context.Context, m *ACMEAccountModel) error {
	info, err := r.client.GetACMEAccountInfo(ctx, m.ACMEType.ValueString())
	if err != nil {
		return err
	}
	m.AccountID = types.StringValue(info.AccountID)
	m.AccountDirectory = types.StringValue(info.AccountDirectory)
	if m.Email.IsNull() && info.AccountEmail != "" {
		m.Email = types.StringValue(info.AccountEmail)
	}
	url, err := r.client.GetACMEDirectoryURL(ctx, m.ACMEType.ValueString())
	if err == nil {
		m.DirectoryURL = types.StringValue(url)
	}
	period, err := r.client.GetACMERenewPeriod(ctx, m.ACMEType.ValueString())
	if err == nil {
		m.RenewPeriod = types.Int64Value(int64(period))
	}
	return nil
}

func (r *ACMEAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ACMEAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.DirectoryURL.IsNull() && !data.DirectoryURL.IsUnknown() && data.DirectoryURL.ValueString() != "" {
		if err := r.client.SetACMEDirectoryURL(ctx, data.ACMEType.ValueString(), data.DirectoryURL.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting ACME directory URL", err.Error())
			return
		}
	}

	if !data.KID.IsNull() && !data.KID.IsUnknown() && data.KID.ValueString() != "" {
		if err := r.client.SetACMEKID(ctx, data.KID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting DigiCert key ID", err.Error())
			return
		}
	}
	if !data.HMACKey.IsNull() && !data.HMACKey.IsUnknown() && data.HMACKey.ValueString() != "" {
		if err := r.client.SetACMEHMAC(ctx, data.HMACKey.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting DigiCert HMAC", err.Error())
			return
		}
	}

	if err := r.client.RegisterACMEAccount(ctx, data.ACMEType.ValueString(), data.Email.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error registering ACME account", err.Error())
		return
	}

	if !data.RenewPeriod.IsNull() && !data.RenewPeriod.IsUnknown() {
		if err := r.client.SetACMERenewPeriod(ctx, data.ACMEType.ValueString(), int32(data.RenewPeriod.ValueInt64())); err != nil {
			resp.Diagnostics.AddError("Error setting ACME renew period", err.Error())
			return
		}
	}

	if err := r.refresh(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading back ACME account state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACMEAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ACMEAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.refresh(ctx, &data); err != nil {
		// LoadMaster doesn't return a clean "no account registered" signal —
		// any error here we treat as resource gone, matching the Read
		// semantics of other resources.
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACMEAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ACMEAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.DirectoryURL.IsNull() && !data.DirectoryURL.IsUnknown() && data.DirectoryURL.ValueString() != "" {
		if err := r.client.SetACMEDirectoryURL(ctx, data.ACMEType.ValueString(), data.DirectoryURL.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error updating ACME directory URL", err.Error())
			return
		}
	}
	if !data.RenewPeriod.IsNull() && !data.RenewPeriod.IsUnknown() {
		if err := r.client.SetACMERenewPeriod(ctx, data.ACMEType.ValueString(), int32(data.RenewPeriod.ValueInt64())); err != nil {
			resp.Diagnostics.AddError("Error updating ACME renew period", err.Error())
			return
		}
	}
	if !data.KID.IsNull() && !data.KID.IsUnknown() && data.KID.ValueString() != "" {
		if err := r.client.SetACMEKID(ctx, data.KID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error updating DigiCert key ID", err.Error())
			return
		}
	}
	if !data.HMACKey.IsNull() && !data.HMACKey.IsUnknown() && data.HMACKey.ValueString() != "" {
		if err := r.client.SetACMEHMAC(ctx, data.HMACKey.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error updating DigiCert HMAC", err.Error())
			return
		}
	}
	if err := r.refresh(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading back ACME account state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACMEAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ACMEAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteACMEConfig(ctx, data.ACMEType.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting ACME config",
			fmt.Sprintf("%s\n\nNote: delacmeconfig fails if any ACME-issued certs still exist for this provider type. Remove kemp_acme_certificate resources first.", err.Error()))
	}
}

// ImportState accepts the acme_type as the import ID ("1" or "2").
func (r *ACMEAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("acme_type"), req.ID)...)
	data := ACMEAccountModel{ACMEType: types.StringValue(req.ID)}
	if err := r.refresh(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing ACME account", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ resource.Resource                = &UserCertResource{}
	_ resource.ResourceWithImportState = &UserCertResource{}
)

func NewUserCertResource() resource.Resource { return &UserCertResource{} }

type UserCertResource struct {
	client *loadmaster.Client
}

type UserCertResourceModel struct {
	Username    types.String `tfsdk:"username"`
	Passphrase  types.String `tfsdk:"passphrase"`
	Certificate types.String `tfsdk:"certificate"`
}

func (r *UserCertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_user_certificate"
}

func (r *UserCertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates and manages a client certificate for a LoadMaster local user account. The certificate can be used for certificate-based admin UI authentication.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Username of the local account for which to generate the certificate. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"passphrase": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Optional. Passphrase to protect the generated private key. Write-only — not read back from the API. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"certificate": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Computed. PEM-encoded certificate and private key for the user. Populated after generation.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *UserCertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserCertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserCertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passphrase := ""
	if !data.Passphrase.IsNull() && !data.Passphrase.IsUnknown() {
		passphrase = data.Passphrase.ValueString()
	}

	if err := r.client.NewUserCert(ctx, data.Username.ValueString(), passphrase); err != nil {
		resp.Diagnostics.AddError("Error generating user certificate", err.Error())
		return
	}

	cert, err := r.client.DownloadUserCert(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user certificate after generation", err.Error())
		return
	}
	data.Certificate = types.StringValue(cert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserCertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserCertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := r.client.DownloadUserCert(ctx, data.Username.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user certificate", err.Error())
		return
	}
	data.Certificate = types.StringValue(cert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserCertResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All mutable fields use RequiresReplace; this is never called.
}

func (r *UserCertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserCertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUserCert(ctx, data.Username.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting user certificate", err.Error())
	}
}

func (r *UserCertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cert, err := r.client.DownloadUserCert(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing user certificate", err.Error())
		return
	}
	data := UserCertResourceModel{
		Username:    types.StringValue(req.ID),
		Passphrase:  types.StringNull(),
		Certificate: types.StringValue(cert),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

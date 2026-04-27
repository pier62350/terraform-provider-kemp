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
	_ resource.Resource                = &CertificateResource{}
	_ resource.ResourceWithImportState = &CertificateResource{}
)

func NewCertificateResource() resource.Resource { return &CertificateResource{} }

type CertificateResource struct {
	client *loadmaster.Client
}

type CertificateResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Data     types.String `tfsdk:"data"`
	Password types.String `tfsdk:"password"`
	Type     types.String `tfsdk:"type"`
}

func (r *CertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an SSL/TLS certificate stored on a Kemp LoadMaster (PEM or PFX). Cert content is immutable; changes force replacement.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Identifier (name) under which LoadMaster stores the certificate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"data": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Base64-encoded PFX bundle or PEM text (cert + key concatenated). Use `base64encode(file(\"...pem\"))` in HCL.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password for encrypted PFX bundles. Leave empty for plain PEM.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cert type as reported by LoadMaster (e.g. `cert`, `pfx`).",
			},
		},
	}
}

func (r *CertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*loadmaster.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *loadmaster.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddCertificate(ctx, data.Name.ValueString(), data.Data.ValueString(), data.Password.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error uploading certificate", err.Error())
		return
	}

	info, err := r.client.FindCertificate(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading certificate after create", err.Error())
		return
	}
	if info == nil {
		resp.Diagnostics.AddError("Certificate disappeared", "addcert succeeded but the certificate is not in listcert output")
		return
	}
	data.Type = types.StringValue(info.Type)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := r.client.FindCertificate(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing certificates", err.Error())
		return
	}
	if info == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Type = types.StringValue(info.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All updatable attributes are ForceNew; this should never be reached.
	resp.Diagnostics.AddError(
		"Update not supported",
		"kemp_certificate has no in-place updatable fields. Changes to name/data/password trigger replacement.",
	)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCertificate(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting certificate", err.Error())
	}
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	info, err := r.client.FindCertificate(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing certificate", err.Error())
		return
	}
	if info == nil {
		resp.Diagnostics.AddError("Certificate not found", fmt.Sprintf("no certificate named %q on the LoadMaster", req.ID))
		return
	}
	// Note: cert "data" cannot be read back from LoadMaster, so the imported
	// state will not include it. A subsequent plan will show a diff if the
	// HCL specifies a `data` attribute.
	data := CertificateResourceModel{
		Name: types.StringValue(info.Name),
		Type: types.StringValue(info.Type),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

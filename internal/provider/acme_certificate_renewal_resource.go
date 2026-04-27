// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource = &ACMECertificateRenewalResource{}
)

func NewACMECertificateRenewalResource() resource.Resource { return &ACMECertificateRenewalResource{} }

type ACMECertificateRenewalResource struct {
	client *loadmaster.Client
}

// ACMECertificateRenewalModel represents a one-shot renewal trigger. The
// resource has no real remote state — Create calls renewacmecert, Read is a
// no-op, Delete is a no-op. Changing triggers or cert_name destroys and
// recreates, which fires another renewal.
type ACMECertificateRenewalModel struct {
	CertName types.String `tfsdk:"cert_name"`
	ACMEType types.String `tfsdk:"acme_type"`
	Triggers types.Map    `tfsdk:"triggers"`
	ID       types.String `tfsdk:"id"`
}

func (r *ACMECertificateRenewalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_certificate_renewal"
}

func (r *ACMECertificateRenewalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Triggers a manual renewal of a ` + "`kemp_acme_certificate`" + ` via ` + "`renewacmecert`" + `.

This is an action resource — Create fires the renewal; Read and Delete are no-ops. Use the ` + "`triggers`" + ` map to force re-renewal: any change to triggers destroys and recreates the resource, firing another renewal.

Note: ` + "`renewacmecert`" + ` is asynchronous. LoadMaster accepts the request and returns immediately; the actual certificate refresh happens in the background.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cert_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the ACME certificate to renew (matches `name` on `kemp_acme_certificate`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"acme_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ACME provider: `letsencrypt` or `digicert`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Arbitrary key-value map. Any change to this map destroys and recreates the resource, triggering another renewal.",
				PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ACMECertificateRenewalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ACMECertificateRenewalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ACMECertificateRenewalModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RenewACMECertificate(ctx, data.CertName.ValueString(), acmeTypeToAPI(data.ACMEType.ValueString())); err != nil {
		resp.Diagnostics.AddError("Error renewing ACME certificate", err.Error())
		return
	}

	data.ID = types.StringValue(data.CertName.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is a no-op: renewacmecert is a one-shot trigger with no queryable state.
func (r *ACMECertificateRenewalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ACMECertificateRenewalModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACMECertificateRenewalResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All meaningful attributes use RequiresReplace, so Update is never called.
}

// Delete is a no-op: there is nothing to undo for a renewal trigger.
func (r *ACMECertificateRenewalResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

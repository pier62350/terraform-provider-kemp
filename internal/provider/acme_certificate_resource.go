// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ACMECertificateResource{}
	_ resource.ResourceWithImportState = &ACMECertificateResource{}
)

func NewACMECertificateResource() resource.Resource { return &ACMECertificateResource{} }

type ACMECertificateResource struct {
	client *loadmaster.Client
}

type ACMECertificateResourceModel struct {
	Name                  types.String `tfsdk:"name"`
	CommonName            types.String `tfsdk:"common_name"`
	VirtualServiceId      types.String `tfsdk:"virtual_service_id"`
	ACMEType              types.String `tfsdk:"acme_type"`
	KeySize               types.Int64  `tfsdk:"key_size"`
	DNSAPI                types.String `tfsdk:"dns_api"`
	DNSAPIParams          types.String `tfsdk:"dns_api_params"`
	Email                 types.String `tfsdk:"email"`
	DomainName            types.String `tfsdk:"domain_name"`
	ExpiryDate            types.String `tfsdk:"expiry_date"`
	Type                  types.String `tfsdk:"type"`
	SubjectAlternateNames types.String `tfsdk:"subject_alternate_names"`
	HTTPChallengeVS       types.String `tfsdk:"http_challenge_vs"`
}

func (r *ACMECertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_certificate"
}

func (r *ACMECertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages an ACME-issued certificate (Let's Encrypt or DigiCert) on a Kemp LoadMaster.

The LoadMaster's ACME service must be configured first (account registered, directory URL set). For Let's Encrypt this is typically done via the WUI under **Certificates & Security → ACME Certificates**.

Issuance is asynchronous: ` + "`addacmecert`" + ` returns immediately, but the cert may take seconds to minutes to be issued. ` + "`terraform apply`" + ` returns once the request is accepted; subsequent ` + "`terraform refresh`" + ` reflects the issued cert.

For wildcard certificates (CN starting with ` + "`*.`" + `), set ` + "`dns_api`" + ` and ` + "`dns_api_params`" + ` so LoadMaster can complete a DNS-01 challenge.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Identifier (name) under which LoadMaster stores the cert.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"common_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Common Name (FQDN) on the certificate. Use a leading `*.` for a wildcard.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"virtual_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Index of the Virtual Service that LoadMaster will use to serve the HTTP-01 challenge. Ignored for DNS-01 (wildcard) issuance but still required by the API.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"acme_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ACME provider: `letsencrypt` (default) or `digicert`.",
				Default:             stringdefault.StaticString("letsencrypt"),
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"key_size": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "RSA key size in bits. LoadMaster default is 2048.",
				Default:             int64default.StaticInt64(2048),
				PlanModifiers:       []planmodifier.Int64{int64RequiresReplace()},
			},
			"dns_api": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "DNS provider for DNS-01 challenge (wildcard certs). Examples: `godaddy.com`, `cloudflare.com`. Required when `common_name` starts with `*.`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"dns_api_params": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Provider-specific credentials for the DNS API (format depends on provider). Required when `dns_api` is set.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Registration email used by the ACME account.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"domain_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Domain name on the issued cert (as reported by LoadMaster).",
			},
			"expiry_date": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Expiry timestamp (LoadMaster's free-form `Mon DD HH:MM:SS YYYY GMT` format).",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cert algorithm (`rsa`, `ecc`).",
			},
			"subject_alternate_names": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Comma-separated SANs as reported by LoadMaster.",
			},
			"http_challenge_vs": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "VS endpoint (IP:port) used for the HTTP-01 challenge.",
			},
		},
	}
}

func (r *ACMECertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ACMECertificateResource) writeState(info *loadmaster.ACMECertificateInfo, m *ACMECertificateResourceModel) {
	m.DomainName = types.StringValue(info.DomainName)
	m.ExpiryDate = types.StringValue(info.ExpiryDate)
	m.Type = types.StringValue(info.Type)
	m.SubjectAlternateNames = types.StringValue(info.SubjectAlternateNames)
	m.HTTPChallengeVS = types.StringValue(info.HTTPChallengeVS)
}

func (r *ACMECertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ACMECertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := loadmaster.AddACMECertParams{
		CommonName:       data.CommonName.ValueString(),
		VirtualServiceID: data.VirtualServiceId.ValueString(),
		ACMEType:         acmeTypeToAPI(data.ACMEType.ValueString()),
		KeySize:          int(data.KeySize.ValueInt64()),
		DNSAPI:           data.DNSAPI.ValueString(),
		DNSAPIParams:     data.DNSAPIParams.ValueString(),
		Email:            data.Email.ValueString(),
	}

	if err := r.client.AddACMECertificate(ctx, data.Name.ValueString(), params); err != nil {
		resp.Diagnostics.AddError("Error requesting ACME certificate", err.Error())
		return
	}

	// Best-effort post-create read; if the cert isn't visible yet (issuance
	// pending), we still persist the configuration so the next refresh can
	// pick it up.
	info, err := r.client.GetACMECertificate(ctx, data.Name.ValueString(), acmeTypeToAPI(data.ACMEType.ValueString()))
	if err == nil {
		r.writeState(info, &data)
	} else {
		data.DomainName = types.StringValue("")
		data.ExpiryDate = types.StringValue("")
		data.Type = types.StringValue("")
		data.SubjectAlternateNames = types.StringValue("")
		data.HTTPChallengeVS = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACMECertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ACMECertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := r.client.GetACMECertificate(ctx, data.Name.ValueString(), acmeTypeToAPI(data.ACMEType.ValueString()))
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading ACME certificate", err.Error())
		return
	}
	r.writeState(info, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACMECertificateResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All user-settable attributes are ForceNew; framework should not invoke
	// Update. Returning an error if it ever does makes the bug visible.
	resp.Diagnostics.AddError(
		"Update not supported",
		"kemp_acme_certificate has no in-place updatable attributes; changes trigger replacement.",
	)
}

func (r *ACMECertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ACMECertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteACMECertificate(ctx, data.Name.ValueString(), acmeTypeToAPI(data.ACMEType.ValueString())); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting ACME certificate", err.Error())
	}
}

// ImportState accepts "<name>" (assumes letsencrypt) or "<name>/<acme_type>"
// where acme_type may be a friendly name ("letsencrypt", "digicert") or the
// legacy numeric form ("1", "2").
func (r *ACMECertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name, acmeType := req.ID, "letsencrypt"
	if i := strings.Index(req.ID, "/"); i != -1 {
		name = req.ID[:i]
		acmeType = acmeTypeFromAPI(acmeTypeToAPI(req.ID[i+1:]))
	}

	info, err := r.client.GetACMECertificate(ctx, name, acmeTypeToAPI(acmeType))
	if err != nil {
		resp.Diagnostics.AddError("Error importing ACME certificate", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("acme_type"), acmeType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("common_name"), info.DomainName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_service_id"), "")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key_size"), parseKeySize(info.KeySize))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_name"), info.DomainName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("expiry_date"), info.ExpiryDate)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), info.Type)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subject_alternate_names"), info.SubjectAlternateNames)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("http_challenge_vs"), info.HTTPChallengeVS)...)
}

func parseKeySize(s string) int64 {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int64(c-'0')
	}
	if n == 0 {
		return 2048
	}
	return n
}

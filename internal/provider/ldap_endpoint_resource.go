// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &LDAPEndpointResource{}
	_ resource.ResourceWithImportState = &LDAPEndpointResource{}
)

func NewLDAPEndpointResource() resource.Resource { return &LDAPEndpointResource{} }

type LDAPEndpointResource struct {
	client *loadmaster.Client
}

type LDAPEndpointResourceModel struct {
	Name                types.String `tfsdk:"name"`
	LDAPType            types.String `tfsdk:"ldap_type"`
	Server              types.String `tfsdk:"server"`
	RevalidationInterval types.Int64  `tfsdk:"revalidation_interval"`
	ReferralCount       types.Int64  `tfsdk:"referral_count"`
}

var ldapTypeFriendlyToAPI = map[string]string{
	"unencrypted": "0",
	"starttls":    "1",
	"ldaps":       "2",
}

func ldapTypeAPIToFriendly(apiVal string) string {
	lower := strings.ToLower(apiVal)
	switch {
	case strings.Contains(lower, "unencrypted"):
		return "unencrypted"
	case strings.Contains(lower, "starttls"):
		return "starttls"
	case strings.Contains(lower, "ldaps"):
		return "ldaps"
	}
	// Fallback: if we get a raw integer string
	switch apiVal {
	case "0":
		return "unencrypted"
	case "1":
		return "starttls"
	case "2":
		return "ldaps"
	}
	return strings.ToLower(apiVal)
}

func (r *LDAPEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_ldap_endpoint"
}

func (r *LDAPEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an LDAP endpoint on a Kemp LoadMaster for authentication.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Name of the LDAP endpoint. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ldap_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Connection type: `unencrypted`, `starttls`, or `ldaps`. Defaults to `unencrypted`.",
			},
			"server": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Space-separated list of LDAP server addresses.",
			},
			"revalidation_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. How often (in seconds) to re-validate LDAP connectivity. Range: 10–86400. Default: 60.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"referral_count": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Number of LDAP referrals to follow. Range: 0–10. Default: 0.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *LDAPEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LDAPEndpointResource) buildParams(data LDAPEndpointResourceModel) loadmaster.LDAPEndpointParams {
	p := loadmaster.LDAPEndpointParams{}

	if !data.LDAPType.IsNull() && !data.LDAPType.IsUnknown() {
		if apiVal, ok := ldapTypeFriendlyToAPI[data.LDAPType.ValueString()]; ok {
			p.LDAPType = apiVal
		}
	}
	if !data.Server.IsNull() && !data.Server.IsUnknown() {
		p.Server = data.Server.ValueString()
	}
	if !data.RevalidationInterval.IsNull() && !data.RevalidationInterval.IsUnknown() {
		v := int32(data.RevalidationInterval.ValueInt64())
		p.VInterval = &v
	}
	if !data.ReferralCount.IsNull() && !data.ReferralCount.IsUnknown() {
		v := int32(data.ReferralCount.ValueInt64())
		p.ReferralCount = &v
	}
	return p
}

func (r *LDAPEndpointResource) writeState(ep *loadmaster.LDAPEndpoint, data *LDAPEndpointResourceModel) {
	data.LDAPType = types.StringValue(ldapTypeAPIToFriendly(ep.LDAPType))
	data.RevalidationInterval = types.Int64Value(int64(ep.VInterval))
	// Timeout from API is available but not exposed — VInterval is our interval field
	// ReferralCount is not returned by ShowLDAPEndpoint; preserve existing state value
	// if already set, otherwise default to 0
	if data.ReferralCount.IsNull() || data.ReferralCount.IsUnknown() {
		data.ReferralCount = types.Int64Value(0)
	}
	// Server is not returned by ShowLDAPEndpoint; preserve existing state value
	if data.Server.IsNull() || data.Server.IsUnknown() {
		data.Server = types.StringValue("")
	}
}

func (r *LDAPEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LDAPEndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := r.buildParams(data)
	if err := r.client.AddLDAPEndpoint(ctx, data.Name.ValueString(), p); err != nil {
		resp.Diagnostics.AddError("Error creating LDAP endpoint", err.Error())
		return
	}

	ep, err := r.client.ShowLDAPEndpoint(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading LDAP endpoint after create", err.Error())
		return
	}
	r.writeState(ep, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LDAPEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LDAPEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ep, err := r.client.ShowLDAPEndpoint(ctx, data.Name.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading LDAP endpoint", err.Error())
		return
	}
	r.writeState(ep, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LDAPEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LDAPEndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := r.buildParams(data)
	if err := r.client.ModifyLDAPEndpoint(ctx, data.Name.ValueString(), p); err != nil {
		resp.Diagnostics.AddError("Error updating LDAP endpoint", err.Error())
		return
	}

	ep, err := r.client.ShowLDAPEndpoint(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading LDAP endpoint after update", err.Error())
		return
	}
	r.writeState(ep, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LDAPEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LDAPEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteLDAPEndpoint(ctx, data.Name.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting LDAP endpoint", err.Error())
	}
}

func (r *LDAPEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ep, err := r.client.ShowLDAPEndpoint(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing LDAP endpoint", err.Error())
		return
	}
	data := LDAPEndpointResourceModel{
		Name:          types.StringValue(req.ID),
		Server:        types.StringValue(""),
		ReferralCount: types.Int64Value(0),
	}
	r.writeState(ep, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ resource.Resource                = &GSLBParamsResource{}
	_ resource.ResourceWithImportState = &GSLBParamsResource{}
)

func NewGSLBParamsResource() resource.Resource { return &GSLBParamsResource{} }

type GSLBParamsResource struct {
	client *loadmaster.Client
}

type GSLBParamsResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Zone              types.String `tfsdk:"zone"`
	SourceOfAuthority types.String `tfsdk:"source_of_authority"`
	Nameserver        types.String `tfsdk:"nameserver"`
	SOAEmail          types.String `tfsdk:"soa_email"`
	TTL               types.Int64  `tfsdk:"ttl"`
	Persist           types.Int64  `tfsdk:"persist"`
}

func (r *GSLBParamsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gslb_params"
}

func (r *GSLBParamsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the GSLB miscellaneous parameters (DNS zone, SOA settings) on a Kemp LoadMaster.

This is a singleton resource. Import with: ` + "`terraform import kemp_gslb_params.main loadmaster`",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `loadmaster`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. DNS zone name for GSLB responses (e.g. `example.com`).",
			},
			"source_of_authority": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Source of authority hostname for the SOA record.",
			},
			"nameserver": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Primary nameserver hostname.",
			},
			"soa_email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Contact email address for the SOA record.",
			},
			"ttl": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Time-to-live for DNS responses, in seconds.",
			},
			"persist": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Persistence timeout in milliseconds.",
			},
		},
	}
}

func (r *GSLBParamsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GSLBParamsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GSLBParamsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetGSLBParams(ctx, r.modelToParams(data)); err != nil {
		resp.Diagnostics.AddError("Error setting GSLB params", err.Error())
		return
	}

	data.ID = types.StringValue("loadmaster")
	if err := r.readInto(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading GSLB params after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBParamsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GSLBParamsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("loadmaster")
	if err := r.readInto(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading GSLB params", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBParamsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GSLBParamsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetGSLBParams(ctx, r.modelToParams(data)); err != nil {
		resp.Diagnostics.AddError("Error updating GSLB params", err.Error())
		return
	}

	data.ID = types.StringValue("loadmaster")
	if err := r.readInto(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading GSLB params after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBParamsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton — removing from state is sufficient; LoadMaster config is unchanged.
}

func (r *GSLBParamsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	data := GSLBParamsResourceModel{ID: types.StringValue("loadmaster")}
	if err := r.readInto(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing GSLB params", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBParamsResource) modelToParams(data GSLBParamsResourceModel) loadmaster.GSLBParams {
	return loadmaster.GSLBParams{
		Zone:              data.Zone.ValueString(),
		SourceOfAuthority: data.SourceOfAuthority.ValueString(),
		Namesrv:           data.Nameserver.ValueString(),
		SOAEmail:          data.SOAEmail.ValueString(),
		TTL:               int32(data.TTL.ValueInt64()),
		Persist:           int32(data.Persist.ValueInt64()),
	}
}

func (r *GSLBParamsResource) readInto(ctx context.Context, data *GSLBParamsResourceModel) error {
	p, err := r.client.GetGSLBParams(ctx)
	if err != nil {
		return err
	}
	data.Zone = types.StringValue(p.Zone)
	data.SourceOfAuthority = types.StringValue(p.SourceOfAuthority)
	data.Nameserver = types.StringValue(p.Namesrv)
	data.SOAEmail = types.StringValue(p.SOAEmail)
	data.TTL = types.Int64Value(int64(p.TTL))
	data.Persist = types.Int64Value(int64(p.Persist))
	return nil
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &GSLBLocationResource{}
	_ resource.ResourceWithImportState = &GSLBLocationResource{}
)

func NewGSLBLocationResource() resource.Resource { return &GSLBLocationResource{} }

type GSLBLocationResource struct {
	client *loadmaster.Client
}

type GSLBLocationResourceModel struct {
	Name types.String `tfsdk:"name"`
}

func (r *GSLBLocationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gslb_location"
}

func (r *GSLBLocationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom GSLB geographic location name on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Location name. Acts as the resource ID. Can be renamed in-place.",
			},
		},
	}
}

func (r *GSLBLocationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GSLBLocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GSLBLocationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddCustomLocation(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating GSLB location", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBLocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GSLBLocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	locations, err := r.client.ListCustomLocations(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading GSLB locations", err.Error())
		return
	}

	name := data.Name.ValueString()
	for _, loc := range locations {
		if loc == name {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *GSLBLocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GSLBLocationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameCustomLocation(ctx, state.Name.ValueString(), plan.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error renaming GSLB location", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GSLBLocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GSLBLocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCustomLocation(ctx, data.Name.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting GSLB location", err.Error())
	}
}

func (r *GSLBLocationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	locations, err := r.client.ListCustomLocations(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing GSLB location", err.Error())
		return
	}

	for _, loc := range locations {
		if loc == req.ID {
			data := GSLBLocationResourceModel{
				Name: types.StringValue(loc),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.Diagnostics.AddError("GSLB location not found", fmt.Sprintf("No GSLB location named %q found.", req.ID))
}

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
	_ resource.Resource                = &RouteResource{}
	_ resource.ResourceWithImportState = &RouteResource{}
)

func NewRouteResource() resource.Resource { return &RouteResource{} }

type RouteResource struct {
	client *loadmaster.Client
}

type RouteResourceModel struct {
	Destination types.String `tfsdk:"destination"`
	Gateway     types.String `tfsdk:"gateway"`
}

func (r *RouteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

func (r *RouteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a static route on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"destination": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Destination network in CIDR notation (e.g. `10.0.0.0/24` or `192.168.1.1/32`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"gateway": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Gateway IP address for the route.",
			},
		},
	}
}

func (r *RouteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddRoute(ctx, data.Destination.ValueString(), data.Gateway.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating route", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	routes, err := r.client.ListRoutes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading routes", err.Error())
		return
	}

	dest := data.Destination.ValueString()
	for _, route := range routes {
		if route.Destination == dest {
			data.Gateway = types.StringValue(route.Gateway)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *RouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRoute(ctx, state.Destination.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting old route during update", err.Error())
		return
	}
	if err := r.client.AddRoute(ctx, data.Destination.ValueString(), data.Gateway.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating route during update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRoute(ctx, data.Destination.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting route", err.Error())
	}
}

func (r *RouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	routes, err := r.client.ListRoutes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing route", err.Error())
		return
	}

	for _, route := range routes {
		if route.Destination == req.ID {
			data := RouteResourceModel{
				Destination: types.StringValue(route.Destination),
				Gateway:     types.StringValue(route.Gateway),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.Diagnostics.AddError("Route not found", fmt.Sprintf("No route with destination %q found.", req.ID))
}

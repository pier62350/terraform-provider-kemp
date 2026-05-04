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
	_ resource.Resource                = &GSLBClusterResource{}
	_ resource.ResourceWithImportState = &GSLBClusterResource{}
)

func NewGSLBClusterResource() resource.Resource { return &GSLBClusterResource{} }

type GSLBClusterResource struct {
	client *loadmaster.Client
}

type GSLBClusterResourceModel struct {
	IP      types.String `tfsdk:"ip"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	LatSecs types.Int64  `tfsdk:"lat_secs"`
	LonSecs types.Int64  `tfsdk:"lon_secs"`
}

func (r *GSLBClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gslb_cluster"
}

func (r *GSLBClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a GSLB remote LoadMaster pool cluster on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** IP address of the cluster. Acts as the resource identity. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Display name of the cluster.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Cluster type (e.g. `remoteLM`). Defaults to empty.",
			},
			"lat_secs": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Latitude in seconds for geographic load balancing.",
			},
			"lon_secs": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Longitude in seconds for geographic load balancing.",
			},
		},
	}
}

func (r *GSLBClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GSLBClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GSLBClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddCluster(ctx, data.IP.ValueString(), data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating GSLB cluster", err.Error())
		return
	}

	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		if err := r.client.ModifyCluster(ctx, data.IP.ValueString(), data.Name.ValueString(), data.Type.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting GSLB cluster type", err.Error())
			return
		}
	}

	if (!data.LatSecs.IsNull() && !data.LatSecs.IsUnknown()) || (!data.LonSecs.IsNull() && !data.LonSecs.IsUnknown()) {
		if err := r.client.ChangeClusterLocation(ctx, data.IP.ValueString(), data.LatSecs.ValueInt64(), data.LonSecs.ValueInt64()); err != nil {
			resp.Diagnostics.AddError("Error setting GSLB cluster location", err.Error())
			return
		}
	}

	cluster, err := r.client.ShowCluster(ctx, data.IP.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading GSLB cluster after create", err.Error())
		return
	}
	data.Name = types.StringValue(cluster.Name)
	data.Type = types.StringValue(cluster.Type)
	data.LatSecs = types.Int64Value(int64(cluster.LatSecs))
	data.LonSecs = types.Int64Value(int64(cluster.LonSecs))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GSLBClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.ShowCluster(ctx, data.IP.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading GSLB cluster", err.Error())
		return
	}

	data.Name = types.StringValue(cluster.Name)
	data.Type = types.StringValue(cluster.Type)
	data.LatSecs = types.Int64Value(int64(cluster.LatSecs))
	data.LonSecs = types.Int64Value(int64(cluster.LonSecs))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GSLBClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Name != state.Name || plan.Type != state.Type {
		if err := r.client.ModifyCluster(ctx, plan.IP.ValueString(), plan.Name.ValueString(), plan.Type.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error updating GSLB cluster", err.Error())
			return
		}
	}

	if plan.LatSecs != state.LatSecs || plan.LonSecs != state.LonSecs {
		if err := r.client.ChangeClusterLocation(ctx, plan.IP.ValueString(), plan.LatSecs.ValueInt64(), plan.LonSecs.ValueInt64()); err != nil {
			resp.Diagnostics.AddError("Error updating GSLB cluster location", err.Error())
			return
		}
	}

	cluster, err := r.client.ShowCluster(ctx, plan.IP.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading GSLB cluster after update", err.Error())
		return
	}
	plan.Name = types.StringValue(cluster.Name)
	plan.Type = types.StringValue(cluster.Type)
	plan.LatSecs = types.Int64Value(int64(cluster.LatSecs))
	plan.LonSecs = types.Int64Value(int64(cluster.LonSecs))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GSLBClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GSLBClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCluster(ctx, data.IP.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting GSLB cluster", err.Error())
	}
}

func (r *GSLBClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cluster, err := r.client.ShowCluster(ctx, req.ID)
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.Diagnostics.AddError("GSLB cluster not found", fmt.Sprintf("No GSLB cluster with IP %q found.", req.ID))
			return
		}
		resp.Diagnostics.AddError("Error importing GSLB cluster", err.Error())
		return
	}
	data := GSLBClusterResourceModel{
		IP:      types.StringValue(req.ID),
		Name:    types.StringValue(cluster.Name),
		Type:    types.StringValue(cluster.Type),
		LatSecs: types.Int64Value(int64(cluster.LatSecs)),
		LonSecs: types.Int64Value(int64(cluster.LonSecs)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

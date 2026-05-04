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
	_ resource.Resource                = &GSLBIPRangeResource{}
	_ resource.ResourceWithImportState = &GSLBIPRangeResource{}
)

func NewGSLBIPRangeResource() resource.Resource { return &GSLBIPRangeResource{} }

type GSLBIPRangeResource struct {
	client *loadmaster.Client
}

type GSLBIPRangeResourceModel struct {
	IP   types.String `tfsdk:"ip"`
	Lat  types.Int64  `tfsdk:"lat"`
	Long types.Int64  `tfsdk:"long"`
}

func (r *GSLBIPRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gslb_ip_range"
}

func (r *GSLBIPRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a client IP address or CIDR registered for GEO/GSLB routing decisions on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** IP address or CIDR (e.g. `203.0.113.0/24`). Acts as the resource identity. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"lat": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional. Latitude in arc-seconds for geographic routing.",
			},
			"long": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional. Longitude in arc-seconds for geographic routing.",
			},
		},
	}
}

func (r *GSLBIPRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GSLBIPRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GSLBIPRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddGSLBIPRange(ctx, data.IP.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating GSLB IP range", err.Error())
		return
	}

	if r.hasLocation(data) {
		if err := r.client.SetGSLBIPLocation(ctx, data.IP.ValueString(), int32(data.Lat.ValueInt64()), int32(data.Long.ValueInt64())); err != nil {
			resp.Diagnostics.AddError("Error setting GSLB IP location", err.Error())
			return
		}
	}

	r.readInto(ctx, &data, resp.Diagnostics.AddError)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBIPRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GSLBIPRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry, err := r.client.ShowGSLBIPRange(ctx, data.IP.ValueString())
	if err != nil {
		if loadmaster.IsGSLBIPNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading GSLB IP range", err.Error())
		return
	}

	r.applyEntry(&data, entry)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBIPRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GSLBIPRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	latChanged := plan.Lat != state.Lat
	longChanged := plan.Long != state.Long

	if latChanged || longChanged {
		if r.hasLocation(plan) {
			if err := r.client.SetGSLBIPLocation(ctx, plan.IP.ValueString(), int32(plan.Lat.ValueInt64()), int32(plan.Long.ValueInt64())); err != nil {
				resp.Diagnostics.AddError("Error updating GSLB IP location", err.Error())
				return
			}
		} else {
			if err := r.client.DeleteGSLBIPLocation(ctx, plan.IP.ValueString()); err != nil {
				resp.Diagnostics.AddError("Error clearing GSLB IP location", err.Error())
				return
			}
		}
	}

	r.readInto(ctx, &plan, resp.Diagnostics.AddError)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GSLBIPRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GSLBIPRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGSLBIPRange(ctx, data.IP.ValueString()); err != nil && !loadmaster.IsGSLBIPNotFound(err) {
		resp.Diagnostics.AddError("Error deleting GSLB IP range", err.Error())
	}
}

func (r *GSLBIPRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	data := GSLBIPRangeResourceModel{IP: types.StringValue(req.ID)}
	entry, err := r.client.ShowGSLBIPRange(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing GSLB IP range", err.Error())
		return
	}
	r.applyEntry(&data, entry)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBIPRangeResource) hasLocation(data GSLBIPRangeResourceModel) bool {
	return !data.Lat.IsNull() || !data.Long.IsNull()
}

func (r *GSLBIPRangeResource) applyEntry(data *GSLBIPRangeResourceModel, entry *loadmaster.GSLBIPRange) {
	data.IP = types.StringValue(entry.IP)
	if entry.Lat != 0 {
		data.Lat = types.Int64Value(int64(entry.Lat))
	} else if data.Lat.IsNull() {
		data.Lat = types.Int64Null()
	}
	if entry.Long != 0 {
		data.Long = types.Int64Value(int64(entry.Long))
	} else if data.Long.IsNull() {
		data.Long = types.Int64Null()
	}
}

func (r *GSLBIPRangeResource) readInto(ctx context.Context, data *GSLBIPRangeResourceModel, addErr func(string, string)) {
	entry, err := r.client.ShowGSLBIPRange(ctx, data.IP.ValueString())
	if err != nil {
		addErr("Error reading GSLB IP range", err.Error())
		return
	}
	r.applyEntry(data, entry)
}

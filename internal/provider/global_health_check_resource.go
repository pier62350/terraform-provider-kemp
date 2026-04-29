// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &GlobalHealthCheckResource{}
	_ resource.ResourceWithImportState = &GlobalHealthCheckResource{}
)

func NewGlobalHealthCheckResource() resource.Resource { return &GlobalHealthCheckResource{} }

type GlobalHealthCheckResource struct {
	client *loadmaster.Client
}

type GlobalHealthCheckResourceModel struct {
	ID            types.String `tfsdk:"id"`
	RetryInterval types.Int64  `tfsdk:"retry_interval"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	RetryCount    types.Int64  `tfsdk:"retry_count"`
}

func (r *GlobalHealthCheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_health_check"
}

func (r *GlobalHealthCheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the global health check settings on a Kemp LoadMaster.

This is a singleton resource — there is exactly one global health configuration per LoadMaster. Destroying this resource removes it from Terraform state but does not change the LoadMaster configuration.

**Note:** The LoadMaster auto-computes ` + "`retry_interval`" + ` as ` + "`retry_count * timeout + 1`" + ` when the other fields are modified. The provider always reads back the actual value after any write, so ` + "`retry_interval`" + ` may differ from what was specified.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `global`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"retry_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Global health check retry interval in seconds. Range: 9–120. Auto-computed by the LoadMaster as `retry_count * timeout + 1`.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Health check timeout in seconds. Range: 1–60.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"retry_count": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Number of consecutive health check failures before marking a server down. Range: 1–15.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *GlobalHealthCheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func globalHealthParams(data GlobalHealthCheckResourceModel) loadmaster.GlobalHealthParams {
	p := loadmaster.GlobalHealthParams{}
	if !data.RetryInterval.IsNull() && !data.RetryInterval.IsUnknown() {
		p.RetryInterval = strconv.FormatInt(data.RetryInterval.ValueInt64(), 10)
	}
	if !data.Timeout.IsNull() && !data.Timeout.IsUnknown() {
		p.Timeout = strconv.FormatInt(data.Timeout.ValueInt64(), 10)
	}
	if !data.RetryCount.IsNull() && !data.RetryCount.IsUnknown() {
		p.RetryCount = strconv.FormatInt(data.RetryCount.ValueInt64(), 10)
	}
	return p
}

func writeGlobalHealthState(gh *loadmaster.GlobalHealth, data *GlobalHealthCheckResourceModel) {
	data.ID = types.StringValue("global")
	if v, err := strconv.ParseInt(gh.RetryInterval, 10, 64); err == nil {
		data.RetryInterval = types.Int64Value(v)
	}
	if v, err := strconv.ParseInt(gh.Timeout, 10, 64); err == nil {
		data.Timeout = types.Int64Value(v)
	}
	if v, err := strconv.ParseInt(gh.RetryCount, 10, 64); err == nil {
		data.RetryCount = types.Int64Value(v)
	}
}

func (r *GlobalHealthCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GlobalHealthCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := globalHealthParams(data)
	if err := r.client.ModifyGlobalHealth(ctx, p); err != nil {
		resp.Diagnostics.AddError("Error configuring global health check", err.Error())
		return
	}

	gh, err := r.client.ShowGlobalHealth(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading global health check after create", err.Error())
		return
	}
	writeGlobalHealthState(gh, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlobalHealthCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GlobalHealthCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gh, err := r.client.ShowGlobalHealth(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading global health check", err.Error())
		return
	}
	writeGlobalHealthState(gh, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlobalHealthCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GlobalHealthCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := globalHealthParams(data)
	if err := r.client.ModifyGlobalHealth(ctx, p); err != nil {
		resp.Diagnostics.AddError("Error updating global health check", err.Error())
		return
	}

	gh, err := r.client.ShowGlobalHealth(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading global health check after update", err.Error())
		return
	}
	writeGlobalHealthState(gh, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlobalHealthCheckResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton — removing from state is sufficient; LoadMaster config is unchanged.
}

func (r *GlobalHealthCheckResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	gh, err := r.client.ShowGlobalHealth(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing global health check", err.Error())
		return
	}
	var data GlobalHealthCheckResourceModel
	writeGlobalHealthState(gh, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

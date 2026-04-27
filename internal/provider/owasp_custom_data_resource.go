// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &OwaspCustomDataResource{}
	_ resource.ResourceWithImportState = &OwaspCustomDataResource{}
)

func NewOwaspCustomDataResource() resource.Resource { return &OwaspCustomDataResource{} }

type OwaspCustomDataResource struct {
	client *loadmaster.Client
}

type OwaspCustomDataModel struct {
	Filename types.String `tfsdk:"filename"`
	Data     types.String `tfsdk:"data"`
}

func (r *OwaspCustomDataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_owasp_custom_data"
}

func (r *OwaspCustomDataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Uploads a custom OWASP data file (admin-level) — typically a word/IP list referenced by ModSecurity rules.",
		Attributes: map[string]schema.Attribute{
			"filename": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Filename including extension (e.g. `owasp_cust.data`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"data": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Base64-encoded data file content.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *OwaspCustomDataResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OwaspCustomDataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OwaspCustomDataModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddOwaspCustomData(ctx, data.Filename.ValueString(), data.Data.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error uploading OWASP custom data file", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OwaspCustomDataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OwaspCustomDataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OwaspCustomDataResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "kemp_owasp_custom_data has no in-place updatable attributes; changes trigger replacement.")
}

func (r *OwaspCustomDataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OwaspCustomDataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOwaspCustomData(ctx, data.Filename.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting OWASP custom data file", err.Error())
	}
}

func (r *OwaspCustomDataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("filename"), req.ID)...)
}

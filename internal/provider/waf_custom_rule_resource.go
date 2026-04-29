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
	_ resource.Resource                = &WafCustomRuleResource{}
	_ resource.ResourceWithImportState = &WafCustomRuleResource{}
)

func NewWafCustomRuleResource() resource.Resource { return &WafCustomRuleResource{} }

type WafCustomRuleResource struct {
	client *loadmaster.Client
}

type WafCustomRuleModel struct {
	Filename types.String `tfsdk:"filename"`
	Data     types.String `tfsdk:"data"`
}

func (r *WafCustomRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waf_custom_rule"
}

func (r *WafCustomRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Uploads a custom rule file for the legacy commercial WAF engine. Distinct from `kemp_owasp_custom_rule`; choose based on which intercept_mode (`1` Legacy vs `2` OWASP) you're configuring on virtual services.",
		Attributes: map[string]schema.Attribute{
			"filename": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Filename — single rule file (e.g. `modsecurity_crs_11_brute_force`) or a tar.gz ruleset bundle. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"data": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "**Required.** Base64-encoded file content. Use `base64encode(file(\"path/to/file\"))`. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *WafCustomRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WafCustomRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WafCustomRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddWafCustomRule(ctx, data.Filename.ValueString(), data.Data.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error uploading WAF custom rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WafCustomRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WafCustomRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WafCustomRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "kemp_waf_custom_rule has no in-place updatable attributes; changes trigger replacement.")
}

func (r *WafCustomRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WafCustomRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWafCustomRule(ctx, data.Filename.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting WAF custom rule", err.Error())
	}
}

func (r *WafCustomRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("filename"), req.ID)...)
}

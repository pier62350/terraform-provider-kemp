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
	_ resource.Resource                = &OwaspCustomRuleResource{}
	_ resource.ResourceWithImportState = &OwaspCustomRuleResource{}
)

func NewOwaspCustomRuleResource() resource.Resource { return &OwaspCustomRuleResource{} }

type OwaspCustomRuleResource struct {
	client *loadmaster.Client
}

type OwaspCustomRuleModel struct {
	Filename types.String `tfsdk:"filename"`
	Data     types.String `tfsdk:"data"`
}

func (r *OwaspCustomRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_owasp_custom_rule"
}

func (r *OwaspCustomRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Uploads a custom OWASP/ModSecurity rule file (admin-level). Once uploaded, the file is referenced by name from `kemp_virtual_service_waf_rule` to attach to specific virtual services.",
		Attributes: map[string]schema.Attribute{
			"filename": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Filename including extension (e.g. `owaspcust.conf`). Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"data": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "**Required.** Base64-encoded rule file content. Use `base64encode(file(\"path/to/rule.conf\"))`. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *OwaspCustomRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OwaspCustomRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OwaspCustomRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddOwaspCustomRule(ctx, data.Filename.ValueString(), data.Data.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error uploading OWASP custom rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OwaspCustomRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// LoadMaster's downloadowaspcustomrule could verify the file exists, but
	// for simplicity we trust state — a missing file will surface on next
	// terraform apply or terraform plan -refresh-only when we add a list cmd.
	var data OwaspCustomRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OwaspCustomRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "kemp_owasp_custom_rule has no in-place updatable attributes; changes trigger replacement.")
}

func (r *OwaspCustomRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OwaspCustomRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOwaspCustomRule(ctx, data.Filename.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting OWASP custom rule", err.Error())
	}
}

func (r *OwaspCustomRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("filename"), req.ID)...)
}

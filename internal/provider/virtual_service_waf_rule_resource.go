// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &VirtualServiceWafRuleResource{}
	_ resource.ResourceWithImportState = &VirtualServiceWafRuleResource{}
)

func NewVirtualServiceWafRuleResource() resource.Resource { return &VirtualServiceWafRuleResource{} }

type VirtualServiceWafRuleResource struct {
	client *loadmaster.Client
}

type VirtualServiceWafRuleModel struct {
	VirtualServiceAddress  types.String `tfsdk:"virtual_service_address"`
	VirtualServicePort     types.String `tfsdk:"virtual_service_port"`
	VirtualServiceProtocol types.String `tfsdk:"virtual_service_protocol"`
	Rule                   types.String `tfsdk:"rule"`
	DisabledRules          types.String `tfsdk:"disabled_rules"`
}

func (r *VirtualServiceWafRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_service_waf_rule"
}

func (r *VirtualServiceWafRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Attaches a WAF rule (or whole rule set) to a virtual service.

LoadMaster's ` + "`vsaddwafrule`" + ` command uses the VS address/port/protocol triplet, not the Index — wire those in from a ` + "`kemp_virtual_service`" + ` resource:

` + "```hcl" + `
resource "kemp_virtual_service_waf_rule" "ipr" {
  virtual_service_address  = kemp_virtual_service.web.address
  virtual_service_port     = kemp_virtual_service.web.port
  virtual_service_protocol = kemp_virtual_service.web.protocol
  rule                     = "G/ip_reputation"
}
` + "```" + `

The ` + "`rule`" + ` value is the LoadMaster rule path (e.g. ` + "`G/ip_reputation`" + `, ` + "`G/malware_detection`" + `). Multiple rules go in one call as a percent-space-separated list (` + "`G/a%20G/b`" + `). Set ` + "`rule`" + ` to empty and provide ` + "`disabled_rules`" + ` to surgically disable individual rule IDs.`,
		Attributes: map[string]schema.Attribute{
			"virtual_service_address":  schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"virtual_service_port":     schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"virtual_service_protocol": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"rule": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "WAF rule path. May be empty if `disabled_rules` is set.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"disabled_rules": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Comma-separated WAF rule IDs to disable (e.g. `2200005,2200006`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *VirtualServiceWafRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtualServiceWafRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtualServiceWafRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddVSWafRule(ctx,
		data.VirtualServiceAddress.ValueString(),
		data.VirtualServicePort.ValueString(),
		data.VirtualServiceProtocol.ValueString(),
		data.Rule.ValueString(),
		data.DisabledRules.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error attaching WAF rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceWafRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// LoadMaster doesn't expose a tidy "list WAF rules attached to this VS"
	// query; the assignment is one-shot. We trust state here. If the rule
	// was removed externally the next apply will re-attach it (idempotent
	// from the user's perspective).
	var data VirtualServiceWafRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualServiceWafRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "kemp_virtual_service_waf_rule has no in-place updatable attributes; changes trigger replacement.")
}

func (r *VirtualServiceWafRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtualServiceWafRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveVSWafRule(ctx,
		data.VirtualServiceAddress.ValueString(),
		data.VirtualServicePort.ValueString(),
		data.VirtualServiceProtocol.ValueString(),
		data.Rule.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error detaching WAF rule", err.Error())
	}
}

// ImportState accepts "<address>/<port>/<protocol>/<rule>".
func (r *VirtualServiceWafRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 4)
	if len(parts) != 4 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf(`expected "<address>/<port>/<protocol>/<rule>", got %q`, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_service_address"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_service_port"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_service_protocol"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule"), parts[3])...)
}

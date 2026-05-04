// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &GSLBFQDNResource{}
	_ resource.ResourceWithImportState = &GSLBFQDNResource{}
)

func NewGSLBFQDNResource() resource.Resource { return &GSLBFQDNResource{} }

type GSLBFQDNResource struct {
	client *loadmaster.Client
}

type GSLBFQDNMemberModel struct {
	IP      types.String `tfsdk:"ip"`
	Cluster types.String `tfsdk:"cluster"`
	Checker types.String `tfsdk:"checker"`
}

type GSLBFQDNResourceModel struct {
	FQDN              types.String          `tfsdk:"fqdn"`
	SelectionCriteria types.String          `tfsdk:"selection_criteria"`
	FailTime          types.Int64           `tfsdk:"fail_time"`
	Members           []GSLBFQDNMemberModel `tfsdk:"member"`
}

func (r *GSLBFQDNResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gslb_fqdn"
}

func (r *GSLBFQDNResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a GSLB FQDN and its IP member list on a Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"fqdn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Fully qualified domain name. Acts as the resource identity. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"selection_criteria": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Load-balancing selection criteria (e.g. `rr`, `wrr`, `lc`, `proximity`, `alwaysfirst`).",
			},
			"fail_time": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Time in milliseconds before a member is marked as failed.",
			},
			"member": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Optional. IP members of this FQDN.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "**Required.** IP address of the member.",
						},
						"cluster": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Optional. Name of the GSLB cluster this member belongs to.",
						},
						"checker": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Optional. Health check method (e.g. `tcp`, `icmp`).",
						},
					},
				},
			},
		},
	}
}

func (r *GSLBFQDNResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GSLBFQDNResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GSLBFQDNResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fqdn := data.FQDN.ValueString()

	if err := r.client.AddFQDN(ctx, fqdn); err != nil {
		resp.Diagnostics.AddError("Error creating GSLB FQDN", err.Error())
		return
	}

	sc := data.SelectionCriteria.ValueString()
	ft := int32(data.FailTime.ValueInt64())
	if sc != "" || ft != 0 {
		if err := r.client.ModifyFQDN(ctx, fqdn, sc, ft); err != nil {
			resp.Diagnostics.AddError("Error configuring GSLB FQDN", err.Error())
			return
		}
	}

	for _, m := range data.Members {
		ip := m.IP.ValueString()
		cluster := m.Cluster.ValueString()
		checker := m.Checker.ValueString()

		if err := r.client.AddFQDNMember(ctx, fqdn, ip, cluster); err != nil {
			resp.Diagnostics.AddError("Error adding GSLB FQDN member", err.Error())
			return
		}
		if checker != "" {
			if err := r.client.ModifyFQDNMember(ctx, fqdn, ip, checker); err != nil {
				resp.Diagnostics.AddError("Error configuring GSLB FQDN member", err.Error())
				return
			}
		}
	}

	if diags := r.refreshState(ctx, &data); diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBFQDNResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GSLBFQDNResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fqdn, err := r.client.ShowFQDN(ctx, data.FQDN.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading GSLB FQDN", err.Error())
		return
	}

	updated := gslbFQDNFromAPI(data.FQDN.ValueString(), fqdn)
	data.SelectionCriteria = updated.SelectionCriteria
	data.FailTime = updated.FailTime
	data.Members = updated.Members
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GSLBFQDNResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GSLBFQDNResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fqdn := plan.FQDN.ValueString()

	if plan.SelectionCriteria != state.SelectionCriteria || plan.FailTime != state.FailTime {
		sc := plan.SelectionCriteria.ValueString()
		ft := int32(plan.FailTime.ValueInt64())
		if err := r.client.ModifyFQDN(ctx, fqdn, sc, ft); err != nil {
			resp.Diagnostics.AddError("Error updating GSLB FQDN", err.Error())
			return
		}
	}

	// Build maps of current and desired members.
	currentMembers := make(map[string]GSLBFQDNMemberModel, len(state.Members))
	for _, m := range state.Members {
		currentMembers[m.IP.ValueString()] = m
	}
	desiredMembers := make(map[string]GSLBFQDNMemberModel, len(plan.Members))
	for _, m := range plan.Members {
		desiredMembers[m.IP.ValueString()] = m
	}

	// Delete removed members.
	for ip := range currentMembers {
		if _, ok := desiredMembers[ip]; !ok {
			if err := r.client.DeleteFQDNMember(ctx, fqdn, ip); err != nil && !loadmaster.IsNotFound(err) {
				resp.Diagnostics.AddError("Error removing GSLB FQDN member", err.Error())
				return
			}
		}
	}

	// Add new members; update changed members.
	for ip, m := range desiredMembers {
		cluster := m.Cluster.ValueString()
		checker := m.Checker.ValueString()

		if _, exists := currentMembers[ip]; !exists {
			// New member.
			if err := r.client.AddFQDNMember(ctx, fqdn, ip, cluster); err != nil {
				resp.Diagnostics.AddError("Error adding GSLB FQDN member", err.Error())
				return
			}
			if checker != "" {
				if err := r.client.ModifyFQDNMember(ctx, fqdn, ip, checker); err != nil {
					resp.Diagnostics.AddError("Error configuring GSLB FQDN member", err.Error())
					return
				}
			}
		} else {
			// Existing member — update if changed.
			cur := currentMembers[ip]
			if m.Checker != cur.Checker {
				if err := r.client.ModifyFQDNMember(ctx, fqdn, ip, checker); err != nil {
					resp.Diagnostics.AddError("Error updating GSLB FQDN member", err.Error())
					return
				}
			}
		}
	}

	if diags := r.refreshState(ctx, &plan); diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GSLBFQDNResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GSLBFQDNResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteFQDN(ctx, data.FQDN.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting GSLB FQDN", err.Error())
	}
}

func (r *GSLBFQDNResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fqdn, err := r.client.ShowFQDN(ctx, req.ID)
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.Diagnostics.AddError("GSLB FQDN not found", fmt.Sprintf("No GSLB FQDN %q found.", req.ID))
			return
		}
		resp.Diagnostics.AddError("Error importing GSLB FQDN", err.Error())
		return
	}

	data := gslbFQDNFromAPI(req.ID, fqdn)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// refreshState fetches the current state from the API and updates data in-place.
// Returns diagnostics on error.
func (r *GSLBFQDNResource) refreshState(ctx context.Context, data *GSLBFQDNResourceModel) diag.Diagnostics {
	fqdn, err := r.client.ShowFQDN(ctx, data.FQDN.ValueString())
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("Error reading GSLB FQDN", err.Error()),
		}
	}
	updated := gslbFQDNFromAPI(data.FQDN.ValueString(), fqdn)
	data.SelectionCriteria = updated.SelectionCriteria
	data.FailTime = updated.FailTime
	data.Members = updated.Members
	return nil
}

func gslbFQDNFromAPI(fqdnStr string, f *loadmaster.GSLBFQDN) GSLBFQDNResourceModel {
	members := make([]GSLBFQDNMemberModel, 0, len(f.Members))
	for _, m := range f.Members {
		members = append(members, GSLBFQDNMemberModel{
			IP:      types.StringValue(m.IP),
			Cluster: types.StringValue(m.Cluster),
			Checker: types.StringValue(m.Checker),
		})
	}
	return GSLBFQDNResourceModel{
		FQDN:              types.StringValue(fqdnStr),
		SelectionCriteria: types.StringValue(f.SelectionCriteria),
		FailTime:          types.Int64Value(int64(f.FailTime)),
		Members:           members,
	}
}

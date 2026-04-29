// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &CipherSetResource{}
	_ resource.ResourceWithImportState = &CipherSetResource{}
)

func NewCipherSetResource() resource.Resource { return &CipherSetResource{} }

type CipherSetResource struct {
	client *loadmaster.Client
}

type CipherSetResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Ciphers types.List   `tfsdk:"ciphers"`
}

func (r *CipherSetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cipher_set"
}

func (r *CipherSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a custom TLS cipher set on a Kemp LoadMaster.

A cipher set is a named, ordered list of TLS cipher strings. Once defined, a cipher set can be assigned to virtual services via the LoadMaster WUI (or via ` + "`modvs`" + ` with the ` + "`CipherSet`" + ` parameter). Built-in sets (` + "`Default`" + `, ` + "`BestPractices`" + `, ` + "`FIPS`" + `, etc.) can be read as data sources but cannot be managed with this resource.

**Cipher order matters** — LoadMaster presents ciphers to clients in the order listed. Put the strongest ciphers first.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "**Required.** Unique name for the cipher set. Forces replacement if changed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ciphers": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "**Required.** Ordered list of OpenSSL cipher strings. LoadMaster joins them with `:` on the wire. Example: `[\"ECDHE-ECDSA-AES256-GCM-SHA384\", \"ECDHE-RSA-AES256-GCM-SHA384\"]`.",
			},
		},
	}
}

func (r *CipherSetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CipherSetResource) ciphersToWire(ctx context.Context, m CipherSetResourceModel) (string, error) {
	var ciphers []string
	if diags := m.Ciphers.ElementsAs(ctx, &ciphers, false); diags.HasError() {
		return "", fmt.Errorf("converting ciphers list")
	}
	return strings.Join(ciphers, ":"), nil
}

func (r *CipherSetResource) writeState(ctx context.Context, cs *loadmaster.CipherSet, m *CipherSetResourceModel) {
	m.Name = types.StringValue(cs.Name)
	var ciphers []string
	if cs.Ciphers != "" {
		for _, c := range strings.Split(cs.Ciphers, ":") {
			c = strings.TrimSpace(c)
			if c != "" {
				ciphers = append(ciphers, c)
			}
		}
	}
	listVal, _ := types.ListValueFrom(ctx, types.StringType, ciphers)
	m.Ciphers = listVal
}

func (r *CipherSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CipherSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wire, err := r.ciphersToWire(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Error building cipher string", err.Error())
		return
	}
	if err := r.client.ModifyCipherSet(ctx, data.Name.ValueString(), wire); err != nil {
		resp.Diagnostics.AddError("Error creating cipher set", err.Error())
		return
	}

	cs, err := r.client.GetCipherSet(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading cipher set after create", err.Error())
		return
	}
	r.writeState(ctx, cs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CipherSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CipherSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cs, err := r.client.GetCipherSet(ctx, data.Name.ValueString())
	if err != nil {
		if loadmaster.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading cipher set", err.Error())
		return
	}
	r.writeState(ctx, cs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CipherSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CipherSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wire, err := r.ciphersToWire(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Error building cipher string", err.Error())
		return
	}
	if err := r.client.ModifyCipherSet(ctx, data.Name.ValueString(), wire); err != nil {
		resp.Diagnostics.AddError("Error updating cipher set", err.Error())
		return
	}

	cs, err := r.client.GetCipherSet(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading cipher set after update", err.Error())
		return
	}
	r.writeState(ctx, cs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CipherSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CipherSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCipherSet(ctx, data.Name.ValueString()); err != nil && !loadmaster.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting cipher set",
			fmt.Sprintf("%s\n\nNote: delcipherset fails if the cipher set is still assigned to any virtual service — unassign it first.", err.Error()))
	}
}

// ImportState accepts the cipher set name as the import ID.
func (r *CipherSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cs, err := r.client.GetCipherSet(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing cipher set", err.Error())
		return
	}
	var data CipherSetResourceModel
	r.writeState(ctx, cs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

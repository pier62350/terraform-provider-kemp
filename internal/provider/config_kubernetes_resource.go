// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var (
	_ resource.Resource                = &ConfigKubernetesResource{}
	_ resource.ResourceWithImportState = &ConfigKubernetesResource{}
)

func NewConfigKubernetesResource() resource.Resource { return &ConfigKubernetesResource{} }

type ConfigKubernetesResource struct {
	client *loadmaster.Client
}

type ConfigKubernetesResourceModel struct {
	ID               types.String `tfsdk:"id"`
	KubeconfigBase64 types.String `tfsdk:"kubeconfig_base64"`
	Mode             types.String `tfsdk:"mode"`
	Namespace        types.String `tfsdk:"namespace"`
	WatchTimeout     types.Int64  `tfsdk:"watch_timeout"`
}

func (r *ConfigKubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_kubernetes"
}

func (r *ConfigKubernetesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the Kemp LoadMaster Kubernetes ingress controller configuration.

This is a singleton resource. Import with: ` + "`terraform import kemp_config_kubernetes.main loadmaster`",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed. Always `loadmaster`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"kubeconfig_base64": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "**Required.** Base64-encoded kubeconfig file content. Use `filebase64(\"~/.kube/config\")` to supply from a file. Sensitive — not echoed in plan output.",
			},
			"mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Ingress controller mode (e.g. `active`, `passive`).",
			},
			"namespace": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Kubernetes namespace to watch for ingress resources.",
			},
			"watch_timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional. Watch timeout in seconds.",
			},
		},
	}
}

func (r *ConfigKubernetesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigKubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigKubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetK8sConfig(ctx, data.KubeconfigBase64.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error uploading kubeconfig", err.Error())
		return
	}

	if !data.Mode.IsNull() && !data.Mode.IsUnknown() {
		if err := r.client.SetK8sMode(ctx, data.Mode.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting Kubernetes ingress mode", err.Error())
			return
		}
	}

	if !data.Namespace.IsNull() && !data.Namespace.IsUnknown() {
		if err := r.client.SetK8sNamespace(ctx, data.Namespace.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting Kubernetes namespace", err.Error())
			return
		}
	}

	if !data.WatchTimeout.IsNull() && !data.WatchTimeout.IsUnknown() {
		if err := r.client.SetK8sWatchTimeout(ctx, strconv.FormatInt(data.WatchTimeout.ValueInt64(), 10)); err != nil {
			resp.Diagnostics.AddError("Error setting Kubernetes watch timeout", err.Error())
			return
		}
	}

	data.ID = types.StringValue("loadmaster")
	if err := r.readRemoteInto(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading Kubernetes config after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigKubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigKubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// kubeconfig_base64 cannot be read back from the API — preserve from state.
	data.ID = types.StringValue("loadmaster")
	if err := r.readRemoteInto(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading Kubernetes config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigKubernetesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ConfigKubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.KubeconfigBase64 != state.KubeconfigBase64 {
		if err := r.client.DeleteK8sConfig(ctx); err != nil {
			resp.Diagnostics.AddError("Error removing old kubeconfig", err.Error())
			return
		}
		if err := r.client.SetK8sConfig(ctx, plan.KubeconfigBase64.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error uploading new kubeconfig", err.Error())
			return
		}
	}

	if plan.Mode != state.Mode && !plan.Mode.IsNull() && !plan.Mode.IsUnknown() {
		if err := r.client.SetK8sMode(ctx, plan.Mode.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error updating Kubernetes ingress mode", err.Error())
			return
		}
	}

	if plan.Namespace != state.Namespace && !plan.Namespace.IsNull() && !plan.Namespace.IsUnknown() {
		if err := r.client.SetK8sNamespace(ctx, plan.Namespace.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error updating Kubernetes namespace", err.Error())
			return
		}
	}

	if plan.WatchTimeout != state.WatchTimeout && !plan.WatchTimeout.IsNull() && !plan.WatchTimeout.IsUnknown() {
		if err := r.client.SetK8sWatchTimeout(ctx, strconv.FormatInt(plan.WatchTimeout.ValueInt64(), 10)); err != nil {
			resp.Diagnostics.AddError("Error updating Kubernetes watch timeout", err.Error())
			return
		}
	}

	plan.ID = types.StringValue("loadmaster")
	if err := r.readRemoteInto(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error reading Kubernetes config after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigKubernetesResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeleteK8sConfig(ctx); err != nil {
		resp.Diagnostics.AddError("Error deleting Kubernetes config", err.Error())
	}
}

func (r *ConfigKubernetesResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// kubeconfig_base64 is write-only and sensitive; start with an empty value.
	// The user must run `terraform apply` to set/reconcile it.
	data := ConfigKubernetesResourceModel{
		ID:               types.StringValue("loadmaster"),
		KubeconfigBase64: types.StringValue(""),
	}
	if err := r.readRemoteInto(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing Kubernetes config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readRemoteInto reads the readable Kubernetes ingress settings (mode, namespace,
// watch_timeout) from the API and sets them on data. kubeconfig_base64 is left
// unchanged (preserved from state).
func (r *ConfigKubernetesResource) readRemoteInto(ctx context.Context, data *ConfigKubernetesResourceModel) error {
	mode, err := r.client.GetK8sMode(ctx)
	if err != nil {
		return fmt.Errorf("GetK8sMode: %w", err)
	}
	data.Mode = types.StringValue(mode)

	ns, err := r.client.GetK8sNamespace(ctx)
	if err != nil {
		return fmt.Errorf("GetK8sNamespace: %w", err)
	}
	data.Namespace = types.StringValue(ns)

	to, err := r.client.GetK8sWatchTimeout(ctx)
	if err != nil {
		return fmt.Errorf("GetK8sWatchTimeout: %w", err)
	}
	if to == "" {
		data.WatchTimeout = types.Int64Value(0)
	} else {
		v, err := strconv.ParseInt(to, 10, 64)
		if err != nil {
			return fmt.Errorf("parse watch_timeout %q: %w", to, err)
		}
		data.WatchTimeout = types.Int64Value(v)
	}
	return nil
}

// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// int64RequiresReplace returns a plan modifier that forces resource
// replacement when the int64 attribute changes. The framework ships a
// stringplanmodifier.RequiresReplace() but the int64 equivalent isn't
// directly importable in older module versions; this is a tiny shim.
func int64RequiresReplace() planmodifier.Int64 {
	return int64RequiresReplaceModifier{}
}

type int64RequiresReplaceModifier struct{}

func (int64RequiresReplaceModifier) Description(context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (int64RequiresReplaceModifier) MarkdownDescription(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (m int64RequiresReplaceModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

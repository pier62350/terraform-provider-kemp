// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"fmt"
)

// AddSubVSRule attaches a content-switching rule to a SubVS.
//
// Kemp's wire command is addrsrule. The parent VS is referenced by its bare
// numeric Index; the SubVS is referenced as "!<subVSIndex>" (bang-prefixed
// Index), which Kemp resolves to the SubVS entry in the parent's pool.
func (c *Client) AddSubVSRule(ctx context.Context, parentVSID, subVSID, ruleName string) error {
	type body struct {
		VS   string `json:"vs"`
		RS   string `json:"rs"`
		Rule string `json:"rule"`
	}
	return c.call(ctx, "addrsrule", body{
		VS:   parentVSID,
		RS:   fmt.Sprintf("!%s", subVSID),
		Rule: ruleName,
	}, nil)
}

// DeleteSubVSRule detaches a content-switching rule from a SubVS.
func (c *Client) DeleteSubVSRule(ctx context.Context, parentVSID, subVSID, ruleName string) error {
	type body struct {
		VS   string `json:"vs"`
		RS   string `json:"rs"`
		Rule string `json:"rule"`
	}
	return c.call(ctx, "delrsrule", body{
		VS:   parentVSID,
		RS:   fmt.Sprintf("!%s", subVSID),
		Rule: ruleName,
	}, nil)
}

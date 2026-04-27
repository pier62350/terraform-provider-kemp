// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// Rule directions for per-VS rule attachment.
const (
	RuleDirectionRequest      = "request"
	RuleDirectionResponse     = "response"
	RuleDirectionResponseBody = "responsebody"
	RuleDirectionPre          = "pre"
)

// AddVSRule attaches an existing system-level rule to a virtual service in
// the given direction. LoadMaster has separate commands per direction
// (addrequestrule / addresponserule / addresponsebodyrule / addprerule);
// the wire fields are otherwise identical.
func (c *Client) AddVSRule(ctx context.Context, vsID, ruleName, direction string) error {
	cmd, err := vsRuleCmd("add", direction)
	if err != nil {
		return err
	}
	type body struct {
		VS   string `json:"vs"`
		Rule string `json:"rule"`
	}
	return c.call(ctx, cmd, body{VS: vsID, Rule: ruleName}, nil)
}

// DeleteVSRule detaches a rule from a VS in the given direction.
func (c *Client) DeleteVSRule(ctx context.Context, vsID, ruleName, direction string) error {
	cmd, err := vsRuleCmd("del", direction)
	if err != nil {
		return err
	}
	type body struct {
		VS   string `json:"vs"`
		Rule string `json:"rule"`
	}
	return c.call(ctx, cmd, body{VS: vsID, Rule: ruleName}, nil)
}

func vsRuleCmd(verb, direction string) (string, error) {
	switch direction {
	case RuleDirectionRequest:
		return verb + "requestrule", nil
	case RuleDirectionResponse:
		return verb + "responserule", nil
	case RuleDirectionResponseBody:
		return verb + "responsebodyrule", nil
	case RuleDirectionPre:
		return verb + "prerule", nil
	}
	return "", &Error{Message: "invalid rule direction: " + direction}
}

// VSHasRule checks whether the named rule is attached to a VS in the given
// direction. Implemented by reading showvs and inspecting the relevant
// MatchRules / RequestRules / ResponseRules / etc. arrays.
//
// Note: showvs returns rule arrays per direction; the exact field names
// observed in production responses are MatchRules (which holds attached
// MatchContentRule names) and similar. The implementation here uses showvs
// and the per-direction arrays it returns, falling back to a not-attached
// signal when LoadMaster's response doesn't include the array yet.
func (c *Client) VSHasRule(ctx context.Context, vsID, ruleName, direction string) (bool, error) {
	type body struct {
		VS string `json:"vs"`
	}
	type vsRules struct {
		Response
		MatchRules         []string `json:"MatchRules,omitempty"`
		RequestRules       []string `json:"RequestRules,omitempty"`
		ResponseRules      []string `json:"ResponseRules,omitempty"`
		ResponseBodyRules  []string `json:"ResponseBodyRules,omitempty"`
		PreProcessRules    []string `json:"PreProcessRules,omitempty"`
	}
	var resp vsRules
	if err := c.call(ctx, "showvs", body{VS: vsID}, &resp); err != nil {
		return false, err
	}
	var list []string
	switch direction {
	case RuleDirectionRequest:
		list = resp.RequestRules
		if len(list) == 0 {
			list = resp.MatchRules
		}
	case RuleDirectionResponse:
		list = resp.ResponseRules
	case RuleDirectionResponseBody:
		list = resp.ResponseBodyRules
	case RuleDirectionPre:
		list = resp.PreProcessRules
	}
	for _, n := range list {
		if n == ruleName {
			return true, nil
		}
	}
	return false, nil
}

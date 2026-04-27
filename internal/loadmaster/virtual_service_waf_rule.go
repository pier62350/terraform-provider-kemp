// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// AddVSWafRule attaches a WAF rule (or rule set) to a virtual service. The
// `rule` parameter is the LoadMaster rule path (e.g. "G/ip_reputation") or
// space-percent-encoded list of multiple paths ("G/a%20G/b"). Pass empty
// rule with non-empty disabledRules to only disable specific rule IDs.
//
// LoadMaster's vsaddwafrule API uses the VS triplet (address/port/protocol),
// not the Index — that's why this method takes three VS-identifying args.
func (c *Client) AddVSWafRule(ctx context.Context, vsAddress, vsPort, vsProtocol, rule, disabledRules string) error {
	type body struct {
		VS           string `json:"vs"`
		Port         string `json:"port"`
		Prot         string `json:"prot"`
		Rule         string `json:"rule"`
		DisableRules string `json:"disablerules,omitempty"`
	}
	return c.call(ctx, "vsaddwafrule", body{
		VS: vsAddress, Port: vsPort, Prot: vsProtocol,
		Rule: rule, DisableRules: disabledRules,
	}, nil)
}

// RemoveVSWafRule detaches a WAF rule from a virtual service.
func (c *Client) RemoveVSWafRule(ctx context.Context, vsAddress, vsPort, vsProtocol, rule string) error {
	type body struct {
		VS   string `json:"vs"`
		Port string `json:"port"`
		Prot string `json:"prot"`
		Rule string `json:"rule"`
	}
	return c.call(ctx, "vsremovewafrule", body{
		VS: vsAddress, Port: vsPort, Prot: vsProtocol, Rule: rule,
	}, nil)
}

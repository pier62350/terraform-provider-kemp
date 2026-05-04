// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// AddRealServerRule attaches an existing system-level rule to a real server.
// The VS is referenced by bare numeric Index; the RS is referenced by its
// address and port (not the bang-prefixed index form, because addrsrule uses
// the address/port tuple to identify the RS).
func (c *Client) AddRealServerRule(ctx context.Context, vsID, port, protocol, rsAddress, rsPort, rule string) error {
	type body struct {
		VS     string `json:"vs"`
		Port   string `json:"port"`
		Prot   string `json:"prot"`
		Rs     string `json:"rs"`
		RSPort string `json:"rsport"`
		Rule   string `json:"rule"`
	}
	return c.call(ctx, "addrsrule", body{VS: vsID, Port: port, Prot: protocol, Rs: rsAddress, RSPort: rsPort, Rule: rule}, nil)
}

// DeleteRealServerRule detaches a rule from a real server.
func (c *Client) DeleteRealServerRule(ctx context.Context, vsID, port, protocol, rsAddress, rsPort, rule string) error {
	type body struct {
		VS     string `json:"vs"`
		Port   string `json:"port"`
		Prot   string `json:"prot"`
		Rs     string `json:"rs"`
		RSPort string `json:"rsport"`
		Rule   string `json:"rule"`
	}
	return c.call(ctx, "delrsrule", body{VS: vsID, Port: port, Prot: protocol, Rs: rsAddress, RSPort: rsPort, Rule: rule}, nil)
}

// rsWithRules extends RealServer with optional rule fields returned by showrs.
type rsWithRules struct {
	Response
	Rs []struct {
		RsIndex int32    `json:"RsIndex"`
		Rules   []string `json:"Rules,omitempty"`
	} `json:"Rs"`
}

// RealServerHasRule checks whether the named rule is attached to a real server
// by calling showrs and inspecting the Rules field. If showrs does not return
// a Rules array (older firmware), the function returns (false, nil) so that
// Terraform will reconcile on the next apply.
func (c *Client) RealServerHasRule(ctx context.Context, vsID, rsID, rule string) (bool, error) {
	type body struct {
		VS string `json:"vs"`
		Rs string `json:"rs"`
	}
	var resp rsWithRules
	if err := c.call(ctx, "showrs", body{VS: vsID, Rs: "!" + rsID}, &resp); err != nil {
		return false, err
	}
	for _, rs := range resp.Rs {
		for _, r := range rs.Rules {
			if r == rule {
				return true, nil
			}
		}
	}
	return false, nil
}

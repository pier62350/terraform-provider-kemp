// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// Rule type enum — values come straight from LoadMaster's modrule documentation.
const (
	RuleTypeMatchContent   = "0"
	RuleTypeAddHeader      = "1"
	RuleTypeDeleteHeader   = "2"
	RuleTypeReplaceHeader  = "3"
	RuleTypeModifyURL      = "4"
	RuleTypeReplaceBody    = "5"
)

// MatchContentRule (type 0) — flexible URL/header/body matcher used as the
// trigger for content-based routing decisions.
type MatchContentRule struct {
	Name            string `json:"Name"`
	Pattern         string `json:"Pattern,omitempty"`
	MatchType       string `json:"MatchType,omitempty"`
	AddHost         bool   `json:"AddHost,omitempty"`
	Negate          bool   `json:"Negate,omitempty"`
	CaseIndependent bool   `json:"CaseIndependent,omitempty"`
	IncludeQuery    bool   `json:"IncludeQuery,omitempty"`
	Header          string `json:"Header,omitempty"`
	MustFail        bool   `json:"MustFail,omitempty"`
	SetOnMatch      int32  `json:"SetOnMatch,omitempty"`
	OnlyOnFlag      int32  `json:"OnlyOnFlag,omitempty"`
	OnlyOnNoFlag    int32  `json:"OnlyOnNoFlag,omitempty"`
}

// AddHeaderRule (type 1) — injects a header on the request.
type AddHeaderRule struct {
	Name         string `json:"Name"`
	Header       string `json:"Header"`
	HeaderValue  string `json:"HeaderValue"`
	OnlyOnFlag   int32  `json:"OnlyOnFlag,omitempty"`
	OnlyOnNoFlag int32  `json:"OnlyOnNoFlag,omitempty"`
}

// DeleteHeaderRule (type 2) — strips headers matching a pattern.
type DeleteHeaderRule struct {
	Name         string `json:"Name"`
	Pattern      string `json:"Pattern"`
	OnlyOnFlag   int32  `json:"OnlyOnFlag,omitempty"`
	OnlyOnNoFlag int32  `json:"OnlyOnNoFlag,omitempty"`
}

// ReplaceHeaderRule (type 3) — substitutes within a specific header.
type ReplaceHeaderRule struct {
	Name         string `json:"Name"`
	Header       string `json:"Header"`
	Pattern      string `json:"Pattern"`
	Replacement  string `json:"Replacement"`
	OnlyOnFlag   int32  `json:"OnlyOnFlag,omitempty"`
	OnlyOnNoFlag int32  `json:"OnlyOnNoFlag,omitempty"`
}

// ModifyURLRule (type 4) — rewrites the URI.
type ModifyURLRule struct {
	Name         string `json:"Name"`
	Pattern      string `json:"Pattern"`
	Replacement  string `json:"Replacement"`
	OnlyOnFlag   int32  `json:"OnlyOnFlag,omitempty"`
	OnlyOnNoFlag int32  `json:"OnlyOnNoFlag,omitempty"`
}

// ReplaceBodyRule (type 5) — substitutes within the response body.
type ReplaceBodyRule struct {
	Name         string `json:"Name"`
	Pattern      string `json:"Pattern"`
	Replacement  string `json:"Replacement"`
	OnlyOnFlag   int32  `json:"OnlyOnFlag,omitempty"`
	OnlyOnNoFlag int32  `json:"OnlyOnNoFlag,omitempty"`
}

// listRulesResponse mirrors showrule, which groups all rules by category.
type listRulesResponse struct {
	Response
	MatchContentRule  []MatchContentRule  `json:"MatchContentRule"`
	AddHeaderRule     []AddHeaderRule     `json:"AddHeaderRule"`
	DeleteHeaderRule  []DeleteHeaderRule  `json:"DeleteHeaderRule"`
	ReplaceHeaderRule []ReplaceHeaderRule `json:"ReplaceHeaderRule"`
	ModifyURLRule     []ModifyURLRule     `json:"ModifyURLRule"`
	ReplaceBodyRule   []ReplaceBodyRule   `json:"ReplaceBodyRule"`
}

// RuleParams carries every possible knob accepted by addrule/modrule. Fields
// the caller leaves empty (or zero, for ints/bools — addrule treats unset and
// zero the same) are dropped via omitempty.
type RuleParams struct {
	Type         string `json:"type,omitempty"`
	Pattern      string `json:"pattern,omitempty"`
	MatchType    string `json:"matchtype,omitempty"`
	IncHost      *bool  `json:"inchost,omitempty"`
	NoCase       *bool  `json:"nocase,omitempty"`
	Negate       *bool  `json:"negate,omitempty"`
	IncQuery     *bool  `json:"incquery,omitempty"`
	Header       string `json:"header,omitempty"`
	Replacement  string `json:"replacement,omitempty"`
	SetOnMatch   *int32 `json:"setonmatch,omitempty"`
	OnlyOnFlag   *int32 `json:"onlyonflag,omitempty"`
	OnlyOnNoFlag *int32 `json:"onlyonnoflag,omitempty"`
	MustFail     *bool  `json:"mustfail,omitempty"`
}

// AddRule creates a new system rule. Type must be one of the RuleType*
// constants (it's a numeric string per LoadMaster's wire format).
func (c *Client) AddRule(ctx context.Context, name string, p RuleParams) error {
	type body struct {
		Name string `json:"name"`
		RuleParams
	}
	return c.call(ctx, "addrule", body{Name: name, RuleParams: p}, nil)
}

// ModifyRule updates an existing rule's parameters. The name is the lookup
// key; type cannot change in-place (replace the resource for that).
func (c *Client) ModifyRule(ctx context.Context, name string, p RuleParams) error {
	type body struct {
		Name string `json:"name"`
		RuleParams
	}
	return c.call(ctx, "modrule", body{Name: name, RuleParams: p}, nil)
}

// DeleteRule removes a system rule by name.
func (c *Client) DeleteRule(ctx context.Context, name string) error {
	type body struct {
		Name string `json:"name"`
	}
	return c.call(ctx, "delrule", body{Name: name}, nil)
}

// listRules pulls every rule in every category. Per-resource Read methods
// then filter by name and category to extract the relevant entry.
func (c *Client) listRules(ctx context.Context) (*listRulesResponse, error) {
	var resp listRulesResponse
	if err := c.call(ctx, "showrule", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindMatchContentRule returns the named MatchContentRule, or nil if absent.
func (c *Client) FindMatchContentRule(ctx context.Context, name string) (*MatchContentRule, error) {
	r, err := c.listRules(ctx)
	if err != nil {
		return nil, err
	}
	for i := range r.MatchContentRule {
		if r.MatchContentRule[i].Name == name {
			return &r.MatchContentRule[i], nil
		}
	}
	return nil, nil
}

// FindAddHeaderRule returns the named AddHeaderRule, or nil if absent.
func (c *Client) FindAddHeaderRule(ctx context.Context, name string) (*AddHeaderRule, error) {
	r, err := c.listRules(ctx)
	if err != nil {
		return nil, err
	}
	for i := range r.AddHeaderRule {
		if r.AddHeaderRule[i].Name == name {
			return &r.AddHeaderRule[i], nil
		}
	}
	return nil, nil
}

// FindDeleteHeaderRule returns the named DeleteHeaderRule, or nil if absent.
func (c *Client) FindDeleteHeaderRule(ctx context.Context, name string) (*DeleteHeaderRule, error) {
	r, err := c.listRules(ctx)
	if err != nil {
		return nil, err
	}
	for i := range r.DeleteHeaderRule {
		if r.DeleteHeaderRule[i].Name == name {
			return &r.DeleteHeaderRule[i], nil
		}
	}
	return nil, nil
}

// FindReplaceHeaderRule returns the named ReplaceHeaderRule, or nil if absent.
func (c *Client) FindReplaceHeaderRule(ctx context.Context, name string) (*ReplaceHeaderRule, error) {
	r, err := c.listRules(ctx)
	if err != nil {
		return nil, err
	}
	for i := range r.ReplaceHeaderRule {
		if r.ReplaceHeaderRule[i].Name == name {
			return &r.ReplaceHeaderRule[i], nil
		}
	}
	return nil, nil
}

// FindModifyURLRule returns the named ModifyURLRule, or nil if absent.
func (c *Client) FindModifyURLRule(ctx context.Context, name string) (*ModifyURLRule, error) {
	r, err := c.listRules(ctx)
	if err != nil {
		return nil, err
	}
	for i := range r.ModifyURLRule {
		if r.ModifyURLRule[i].Name == name {
			return &r.ModifyURLRule[i], nil
		}
	}
	return nil, nil
}

// FindReplaceBodyRule returns the named ReplaceBodyRule, or nil if absent.
func (c *Client) FindReplaceBodyRule(ctx context.Context, name string) (*ReplaceBodyRule, error) {
	r, err := c.listRules(ctx)
	if err != nil {
		return nil, err
	}
	for i := range r.ReplaceBodyRule {
		if r.ReplaceBodyRule[i].Name == name {
			return &r.ReplaceBodyRule[i], nil
		}
	}
	return nil, nil
}

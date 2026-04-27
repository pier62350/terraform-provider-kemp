// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// AddOwaspCustomRule uploads a custom OWASP rule file (typically a .conf
// containing ModSecurity rule definitions). data must already be base64-
// encoded.
func (c *Client) AddOwaspCustomRule(ctx context.Context, filename, base64Data string) error {
	type body struct {
		Filename string `json:"filename"`
		Data     string `json:"data"`
	}
	return c.call(ctx, "addowaspcustomrule", body{Filename: filename, Data: base64Data}, nil)
}

// DeleteOwaspCustomRule removes a custom OWASP rule file by name (without
// the extension, per the API examples).
func (c *Client) DeleteOwaspCustomRule(ctx context.Context, filename string) error {
	type body struct {
		Filename string `json:"filename"`
	}
	return c.call(ctx, "delowaspcustomrule", body{Filename: filename}, nil)
}

// AddOwaspCustomData uploads a custom OWASP data file (e.g. word lists used
// by ModSecurity rules). data must already be base64-encoded.
func (c *Client) AddOwaspCustomData(ctx context.Context, filename, base64Data string) error {
	type body struct {
		Filename string `json:"filename"`
		Data     string `json:"data"`
	}
	return c.call(ctx, "addowaspcustomdata", body{Filename: filename, Data: base64Data}, nil)
}

// DeleteOwaspCustomData removes a custom OWASP data file.
func (c *Client) DeleteOwaspCustomData(ctx context.Context, filename string) error {
	type body struct {
		Filename string `json:"filename"`
	}
	return c.call(ctx, "delowaspcustomdata", body{Filename: filename}, nil)
}

// AddWafCustomRule uploads a commercial-WAF custom rule file (the legacy
// WAF engine, separate from OWASP). data must be base64-encoded.
func (c *Client) AddWafCustomRule(ctx context.Context, filename, base64Data string) error {
	type body struct {
		Filename string `json:"filename"`
		Data     string `json:"data"`
	}
	return c.call(ctx, "addwafcustomrule", body{Filename: filename, Data: base64Data}, nil)
}

// DeleteWafCustomRule removes a commercial-WAF custom rule file.
func (c *Client) DeleteWafCustomRule(ctx context.Context, filename string) error {
	type body struct {
		Filename string `json:"filename"`
	}
	return c.call(ctx, "delwafcustomrule", body{Filename: filename}, nil)
}

// AddWafCustomData uploads a commercial-WAF custom data file.
func (c *Client) AddWafCustomData(ctx context.Context, filename, base64Data string) error {
	type body struct {
		Filename string `json:"filename"`
		Data     string `json:"data"`
	}
	return c.call(ctx, "addwafcustomdata", body{Filename: filename, Data: base64Data}, nil)
}

// DeleteWafCustomData removes a commercial-WAF custom data file.
func (c *Client) DeleteWafCustomData(ctx context.Context, filename string) error {
	type body struct {
		Filename string `json:"filename"`
	}
	return c.call(ctx, "delwafcustomdata", body{Filename: filename}, nil)
}

// VerifyOwaspCustomRule confirms a custom OWASP rule file exists on the
// LoadMaster by issuing a download and discarding the content.
func (c *Client) VerifyOwaspCustomRule(ctx context.Context, filename string) error {
	type body struct {
		Filename string `json:"filename"`
	}
	return c.call(ctx, "downloadowaspcustomrule", body{Filename: filename}, nil)
}

// VerifyOwaspCustomData confirms a custom OWASP data file exists on the
// LoadMaster by issuing a download and discarding the content.
func (c *Client) VerifyOwaspCustomData(ctx context.Context, filename string) error {
	type body struct {
		Filename string `json:"filename"`
	}
	return c.call(ctx, "downloadowaspcustomdata", body{Filename: filename}, nil)
}

// VerifyWafCustomRule confirms a commercial-WAF custom rule file exists on the
// LoadMaster.
func (c *Client) VerifyWafCustomRule(ctx context.Context, filename string) error {
	type body struct {
		Filename string `json:"filename"`
	}
	return c.call(ctx, "downloadwafcustomrule", body{Filename: filename}, nil)
}

// VerifyWafCustomData confirms a commercial-WAF custom data file exists on the
// LoadMaster.
func (c *Client) VerifyWafCustomData(ctx context.Context, filename string) error {
	type body struct {
		Filename string `json:"filename"`
	}
	return c.call(ctx, "downloadwafcustomdata", body{Filename: filename}, nil)
}

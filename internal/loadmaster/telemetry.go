// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

type telemetryResponse struct {
	Response
	// The API may return the enabled state at the top level or nested inside an
	// Interface array. Both shapes are handled here.
	TelemetryEnabled any       `json:"TelemetryEnabled,omitempty"`
	Enabled          any       `json:"Enabled,omitempty"`
	Interface        []ifState `json:"Interface,omitempty"`
}

type ifState struct {
	Id               int  `json:"Id"`
	TelemetryEnabled bool `json:"TelemetryEnabled"`
	Enabled          bool `json:"Enabled"`
}

// SetTelemetry enables or disables telemetry on the given interface.
func (c *Client) SetTelemetry(ctx context.Context, interfaceID string, enabled bool) error {
	enableStr := "0"
	if enabled {
		enableStr = "1"
	}
	type body struct {
		Interface string `json:"interface"`
		Enable    string `json:"enable"`
	}
	return c.call(ctx, "enabletelemetry", body{Interface: interfaceID, Enable: enableStr}, nil)
}

// GetTelemetry returns the telemetry enabled state for the given interface.
func (c *Client) GetTelemetry(ctx context.Context, interfaceID string) (bool, error) {
	type body struct {
		Interface string `json:"interface"`
	}
	var resp telemetryResponse
	if err := c.call(ctx, "showtelemetry", body{Interface: interfaceID}, &resp); err != nil {
		return false, err
	}

	// Try the Interface array first (list response shape).
	if len(resp.Interface) > 0 {
		return resp.Interface[0].TelemetryEnabled || resp.Interface[0].Enabled, nil
	}

	// Fall back to top-level fields. The API may return "1"/"0" or true/false.
	return parseBoolAny(resp.TelemetryEnabled) || parseBoolAny(resp.Enabled), nil
}

// parseBoolAny converts JSON bool or string "1"/"0"/"true"/"false" to bool.
func parseBoolAny(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case string:
		return t == "1" || t == "true"
	}
	return false
}

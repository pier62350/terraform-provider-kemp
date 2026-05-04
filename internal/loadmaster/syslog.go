// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"strings"
)

// titleCase uppercases the first letter of s.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// GetSyslogHosts returns the comma-separated list of hosts configured for the
// given syslog severity level. The param is e.g. "SyslogNotice".
func (c *Client) GetSyslogHosts(ctx context.Context, level string) ([]string, error) {
	param := "Syslog" + titleCase(strings.ToLower(level))
	value, err := c.GetParam(ctx, param)
	if err != nil {
		return nil, err
	}
	if value == "" {
		return []string{}, nil
	}
	return strings.Split(value, ","), nil
}

// SetSyslogHosts replaces all hosts for the given severity level.
// Pass an empty slice to clear all hosts for that level.
func (c *Client) SetSyslogHosts(ctx context.Context, level string, hosts []string) error {
	param := "syslog" + strings.ToLower(level)
	return c.SetParam(ctx, param, strings.Join(hosts, ","))
}

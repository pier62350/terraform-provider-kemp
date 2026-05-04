// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"fmt"
)

// WAFSettings holds the WAF auto-update and remote logging configuration
// returned by getwafsettings.
type WAFSettings struct {
	AutoUpdate           *bool   `json:"AutoUpdate,omitempty"`
	AutoInstall          *bool   `json:"AutoInstall,omitempty"`
	InstallHour          *int32  `json:"InstallHour,omitempty"`
	RemoteLoggingEnabled *bool   `json:"RemoteLoggingEnabled,omitempty"`
	RemoteURI            string  `json:"RemoteURI,omitempty"`
	LogFormat            string  `json:"LogFormat,omitempty"`
}

type wafSettingsResponse struct {
	Response
	WAFSettings
}

// GetWAFSettings retrieves the current WAF settings from the LoadMaster.
func (c *Client) GetWAFSettings(ctx context.Context) (*WAFSettings, error) {
	var resp wafSettingsResponse
	if err := c.call(ctx, "getwafsettings", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.WAFSettings, nil
}

// SetWAFAutoUpdate enables or disables WAF automatic rule updates.
func (c *Client) SetWAFAutoUpdate(ctx context.Context, enable bool) error {
	enableStr := "0"
	if enable {
		enableStr = "1"
	}
	type body struct {
		Enable string `json:"enable"`
	}
	return c.call(ctx, "setwafautoupdate", body{Enable: enableStr}, nil)
}

// SetWAFAutoInstall enables or disables automatic installation of WAF rule updates.
func (c *Client) SetWAFAutoInstall(ctx context.Context, enable bool) error {
	enableStr := "0"
	if enable {
		enableStr = "1"
	}
	type body struct {
		Enable string `json:"enable"`
	}
	return c.call(ctx, "setwafautoinstall", body{Enable: enableStr}, nil)
}

// SetWAFInstallHour sets the hour (0–23) at which WAF rules are installed.
func (c *Client) SetWAFInstallHour(ctx context.Context, hour int64) error {
	type body struct {
		Hour string `json:"hour"`
	}
	return c.call(ctx, "setwafinstalltime", body{Hour: fmt.Sprintf("%d", hour)}, nil)
}

// EnableWAFRemoteLogging configures and enables WAF remote logging.
func (c *Client) EnableWAFRemoteLogging(ctx context.Context, remoteURI, username, password string) error {
	type body struct {
		RemoteURI string `json:"remoteuri"`
		Username  string `json:"username"`
		Passwd    string `json:"passwd"`
	}
	return c.call(ctx, "enablewafremotelogging", body{RemoteURI: remoteURI, Username: username, Passwd: password}, nil)
}

// DisableWAFRemoteLogging disables WAF remote logging.
func (c *Client) DisableWAFRemoteLogging(ctx context.Context) error {
	return c.call(ctx, "disablewafremotelogging", nil, nil)
}

// SetWAFLogFormat sets the WAF log format (e.g. "cef", "leef", "w3c").
func (c *Client) SetWAFLogFormat(ctx context.Context, format string) error {
	type body struct {
		LogFormat string `json:"logformat"`
	}
	return c.call(ctx, "setwaflogformat", body{LogFormat: format}, nil)
}

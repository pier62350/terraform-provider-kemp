// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import "context"

// VirtualService mirrors the relevant subset of fields returned by the
// addvs/showvs/modvs commands. Field names match the LoadMaster JSON
// (PascalCase) via explicit json tags.
type VirtualService struct {
	Index           int32  `json:"Index"`
	Address         string `json:"VSAddress"`
	Port            string `json:"VSPort"`
	Protocol        string `json:"Protocol"`
	VSType          string `json:"VStype"`
	NickName        string `json:"NickName"`
	Enable          *bool  `json:"Enable,omitempty"`
	SSLAcceleration *bool  `json:"SSLAcceleration,omitempty"`
	CertFile        string `json:"CertFile,omitempty"`
	CipherSet       string `json:"CipherSet,omitempty"`
	TlsType         string `json:"TlsType,omitempty"`

	// Standard options
	Schedule            string `json:"Schedule,omitempty"`
	Persist             string `json:"Persist,omitempty"`
	PersistTimeout      string `json:"PersistTimeout,omitempty"`
	Idletime            *int32 `json:"Idletime,omitempty"`
	ServerInit          *int32 `json:"ServerInit,omitempty"`
	ForceL7             *bool  `json:"ForceL7,omitempty"`
	ForceL4             *bool  `json:"ForceL4,omitempty"`
	Transparent         *bool  `json:"Transparent,omitempty"`
	UseforSnat          *bool  `json:"UseforSnat,omitempty"`
	MultiConnect        *bool  `json:"MultiConnect,omitempty"`
	Cache               *bool  `json:"Cache,omitempty"`
	Compress            *bool  `json:"Compress,omitempty"`
	AllowHTTP2          *bool  `json:"AllowHTTP2,omitempty"`
	SSLReverse          *bool  `json:"SSLReverse,omitempty"`
	SSLReencrypt        *bool  `json:"SSLReencrypt,omitempty"`
	PassSni             *bool  `json:"PassSni,omitempty"`
	PassCipher          *bool  `json:"PassCipher,omitempty"`
	Verify              *int32 `json:"Verify,omitempty"`
	ClientCert          *int32 `json:"ClientCert,omitempty"`
	AddVia              *int32 `json:"AddVia,omitempty"`
	RefreshPersist      *bool  `json:"RefreshPersist,omitempty"`
	RsMinimum           *int32 `json:"RsMinimum,omitempty"`
	Bandwidth           *int32 `json:"Bandwidth,omitempty"`
	ConnsPerSecLimit    *int32 `json:"ConnsPerSecLimit,omitempty"`
	RequestsPerSecLimit *int32 `json:"RequestsPerSecLimit,omitempty"`
	MaxConnsLimit       *int32 `json:"MaxConnsLimit,omitempty"`

	// Health checks
	CheckType            string `json:"CheckType,omitempty"`
	CheckPort            string `json:"CheckPort,omitempty"`
	ChkInterval          *int32 `json:"ChkInterval,omitempty"`
	ChkTimeout           *int32 `json:"ChkTimeout,omitempty"`
	ChkRetryCount        *int32 `json:"ChkRetryCount,omitempty"`
	NeedHostName         *bool  `json:"NeedHostName,omitempty"`
	CheckUseHTTP11       *bool  `json:"CheckUse1.1,omitempty"`
	CheckUseGet          *int32 `json:"CheckUseGet,omitempty"`
	MatchLen             *int32 `json:"MatchLen,omitempty"`
	EnhancedHealthChecks *bool  `json:"EnhancedHealthChecks,omitempty"`

	// ESP (Edge Security Pack)
	EspEnabled          *bool  `json:"EspEnabled,omitempty"`
	AllowedHosts        string `json:"AllowedHosts,omitempty"`
	AllowedDirectories  string `json:"AllowedDirectories,omitempty"`
	InputAuthMode       string `json:"InputAuthMode,omitempty"`
	OutputAuthMode      string `json:"OutputAuthMode,omitempty"`
	IncludeNestedGroups *bool  `json:"IncludeNestedGroups,omitempty"`
	DisplayPubPriv      *bool  `json:"DisplayPubPriv,omitempty"`
	EspLogs             *bool  `json:"EspLogs,omitempty"`

	// WAF
	InterceptMode        string `json:"InterceptMode,omitempty"`
	BlockingParanoia     *int32 `json:"BlockingParanoia,omitempty"`
	AlertThreshold       *int32 `json:"AlertThreshold,omitempty"`
	IPReputationBlocking *bool  `json:"IPReputationBlocking,omitempty"`
}

// VirtualServiceParams are the optional knobs for create/modify.
// Only fields that are non-nil / non-empty get sent to LoadMaster.
type VirtualServiceParams struct {
	NickName        string `json:"NickName,omitempty"`
	VSType          string `json:"VStype,omitempty"`
	Enable          *bool  `json:"Enable,omitempty"`
	SSLAcceleration *bool  `json:"SSLAcceleration,omitempty"`
	CertFile        string `json:"CertFile,omitempty"`
	CipherSet       string `json:"CipherSet,omitempty"`
	TlsType         string `json:"TlsType,omitempty"`

	// Standard options
	Schedule            string `json:"Schedule,omitempty"`
	Persist             string `json:"persist,omitempty"` // lowercase per wire format
	PersistTimeout      string `json:"PersistTimeout,omitempty"`
	Idletime            *int32 `json:"Idletime,omitempty"`
	ServerInit          *int32 `json:"ServerInit,omitempty"`
	ForceL7             *bool  `json:"ForceL7,omitempty"`
	ForceL4             *bool  `json:"ForceL4,omitempty"`
	Transparent         *bool  `json:"Transparent,omitempty"`
	UseforSnat          *bool  `json:"UseforSnat,omitempty"`
	MultiConnect        *bool  `json:"MultiConnect,omitempty"`
	Cache               *bool  `json:"Cache,omitempty"`
	Compress            *bool  `json:"Compress,omitempty"`
	AllowHTTP2          *bool  `json:"AllowHTTP2,omitempty"`
	SSLReverse          *bool  `json:"SSLReverse,omitempty"`
	SSLReencrypt        *bool  `json:"SSLReencrypt,omitempty"`
	PassSni             *bool  `json:"PassSni,omitempty"`
	PassCipher          *bool  `json:"PassCipher,omitempty"`
	Verify              *int32 `json:"Verify,omitempty"`
	ClientCert          *int32 `json:"ClientCert,omitempty"`
	AddVia              *int32 `json:"AddVia,omitempty"`
	RefreshPersist      *bool  `json:"RefreshPersist,omitempty"`
	RsMinimum           *int32 `json:"RsMinimum,omitempty"`
	Bandwidth           *int32 `json:"Bandwidth,omitempty"`
	ConnsPerSecLimit    *int32 `json:"ConnsPerSecLimit,omitempty"`
	RequestsPerSecLimit *int32 `json:"RequestsPerSecLimit,omitempty"`
	MaxConnsLimit       *int32 `json:"MaxConnsLimit,omitempty"`

	// Health checks
	CheckType            string `json:"CheckType,omitempty"`
	CheckPort            string `json:"CheckPort,omitempty"`
	ChkInterval          *int32 `json:"ChkInterval,omitempty"`
	ChkTimeout           *int32 `json:"ChkTimeout,omitempty"`
	ChkRetryCount        *int32 `json:"ChkRetryCount,omitempty"`
	NeedHostName         *bool  `json:"NeedHostName,omitempty"`
	CheckUseHTTP11       *bool  `json:"CheckUse1.1,omitempty"`
	CheckUseGet          *int32 `json:"CheckUseGet,omitempty"`
	MatchLen             *int32 `json:"MatchLen,omitempty"`
	EnhancedHealthChecks *bool  `json:"EnhancedHealthChecks,omitempty"`

	// ESP (Edge Security Pack)
	EspEnabled          *bool  `json:"EspEnabled,omitempty"`
	AllowedHosts        string `json:"AllowedHosts,omitempty"`
	AllowedDirectories  string `json:"AllowedDirectories,omitempty"`
	InputAuthMode       string `json:"InputAuthMode,omitempty"`
	OutputAuthMode      string `json:"OutputAuthMode,omitempty"`
	IncludeNestedGroups *bool  `json:"IncludeNestedGroups,omitempty"`
	DisplayPubPriv      *bool  `json:"DisplayPubPriv,omitempty"`
	EspLogs             *bool  `json:"EspLogs,omitempty"`

	// WAF
	InterceptMode        string `json:"InterceptMode,omitempty"`
	BlockingParanoia     *int32 `json:"BlockingParanoia,omitempty"`
	AlertThreshold       *int32 `json:"AlertThreshold,omitempty"`
	IPReputationBlocking *bool  `json:"IPReputationBlocking,omitempty"`
}

type vsResponse struct {
	Response
	VirtualService
}

// AddVirtualService creates a new virtual service.
func (c *Client) AddVirtualService(ctx context.Context, address, port, protocol string, p VirtualServiceParams) (*VirtualService, error) {
	type body struct {
		VS       string `json:"vs"`
		Port     string `json:"port"`
		Protocol string `json:"prot"`
		VirtualServiceParams
	}
	var resp vsResponse
	if err := c.call(ctx, "addvs", body{VS: address, Port: port, Protocol: protocol, VirtualServiceParams: p}, &resp); err != nil {
		return nil, err
	}
	return &resp.VirtualService, nil
}

// ShowVirtualService reads a single VS by its numeric Index.
//
// LoadMaster's parser interprets a numeric `vs` as the Index automatically;
// the "!N" prefix syntax is rejected by some firmware revs (it falls back to
// address-mode parsing and errors on missing `port`). Bare numeric Index is
// the safe form across versions.
func (c *Client) ShowVirtualService(ctx context.Context, id string) (*VirtualService, error) {
	type body struct {
		VS string `json:"vs"`
	}
	var resp vsResponse
	if err := c.call(ctx, "showvs", body{VS: id}, &resp); err != nil {
		return nil, err
	}
	return &resp.VirtualService, nil
}

// ModifyVirtualService updates a VS in place.
func (c *Client) ModifyVirtualService(ctx context.Context, id string, p VirtualServiceParams) (*VirtualService, error) {
	type body struct {
		VS string `json:"vs"`
		VirtualServiceParams
	}
	var resp vsResponse
	if err := c.call(ctx, "modvs", body{VS: id, VirtualServiceParams: p}, &resp); err != nil {
		return nil, err
	}
	return &resp.VirtualService, nil
}

// DeleteVirtualService removes a VS by Index.
func (c *Client) DeleteVirtualService(ctx context.Context, id string) error {
	type body struct {
		VS string `json:"vs"`
	}
	return c.call(ctx, "delvs", body{VS: id}, nil)
}

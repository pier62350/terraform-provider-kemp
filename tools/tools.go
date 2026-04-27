// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

//go:build tools

// Package tools tracks the codegen binaries we depend on, so they end up in
// go.sum and are reproducibly invoked via `go run`.
package tools

import (
	// docs generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name kemp

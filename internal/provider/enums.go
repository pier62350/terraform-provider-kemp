// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

// acmeTypeToAPI converts a friendly acme_type value to the numeric string the
// API expects. Unknown values pass through so the API can return a clear error.
func acmeTypeToAPI(s string) string {
	switch s {
	case "letsencrypt":
		return "1"
	case "digicert":
		return "2"
	default:
		return s
	}
}

// acmeTypeFromAPI converts a numeric acme_type returned by the API to its
// friendly name. Numeric pass-through preserves backward compatibility.
func acmeTypeFromAPI(s string) string {
	switch s {
	case "1":
		return "letsencrypt"
	case "2":
		return "digicert"
	default:
		return s
	}
}

// wafInterceptModeToAPI converts a friendly waf_intercept_mode to its numeric
// API value.
func wafInterceptModeToAPI(s string) string {
	switch s {
	case "disabled":
		return "0"
	case "legacy":
		return "1"
	case "owasp":
		return "2"
	default:
		return s
	}
}

// wafInterceptModeFromAPI converts a numeric InterceptMode from the API to its
// friendly name.
func wafInterceptModeFromAPI(s string) string {
	switch s {
	case "0":
		return "disabled"
	case "1":
		return "legacy"
	case "2":
		return "owasp"
	default:
		return s
	}
}

// espInputAuthModeToAPI converts a friendly esp_input_auth_mode value to its
// numeric API string. Known values: "none", "basic", "form".
func espInputAuthModeToAPI(s string) string {
	switch s {
	case "none":
		return "0"
	case "basic":
		return "1"
	case "form":
		return "2"
	default:
		return s
	}
}

// espInputAuthModeFromAPI converts a numeric InputAuthMode from the API to its
// friendly name.
func espInputAuthModeFromAPI(s string) string {
	switch s {
	case "0":
		return "none"
	case "1":
		return "basic"
	case "2":
		return "form"
	default:
		return s
	}
}

// espOutputAuthModeToAPI converts a friendly esp_output_auth_mode value to its
// numeric API string. Known values: "none", "basic", "form", "kcd".
func espOutputAuthModeToAPI(s string) string {
	switch s {
	case "none":
		return "0"
	case "basic":
		return "1"
	case "form":
		return "2"
	case "kcd":
		return "4"
	default:
		return s
	}
}

// espOutputAuthModeFromAPI converts a numeric OutputAuthMode from the API to
// its friendly name.
func espOutputAuthModeFromAPI(s string) string {
	switch s {
	case "0":
		return "none"
	case "1":
		return "basic"
	case "2":
		return "form"
	case "4":
		return "kcd"
	default:
		return s
	}
}

// addViaToAPI converts a friendly add_via value to the *int32 the API expects.
// Values: "no" (0), "add" (1), "replace" (2).
func addViaToAPI(s string) *int32 {
	var v int32
	switch s {
	case "add":
		v = 1
	case "replace":
		v = 2
	default: // "no" or empty
		v = 0
	}
	return &v
}

// addViaFromAPI converts the *int32 AddVia from the API to its friendly name.
func addViaFromAPI(v *int32) string {
	if v == nil {
		return "no"
	}
	switch *v {
	case 1:
		return "add"
	case 2:
		return "replace"
	default:
		return "no"
	}
}

// checkUseGetToAPI converts a friendly check_use_get value to the *int32 the
// API expects. Unknown values default to HEAD (0).
func checkUseGetToAPI(s string) *int32 {
	var v int32
	switch s {
	case "get":
		v = 1
	default: // "head" or empty
		v = 0
	}
	return &v
}

// checkUseGetFromAPI converts the *int32 CheckUseGet from the API to its
// friendly name.
func checkUseGetFromAPI(v *int32) string {
	if v != nil && *v == 1 {
		return "get"
	}
	return "head"
}

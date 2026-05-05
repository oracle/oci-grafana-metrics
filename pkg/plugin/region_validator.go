// Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package plugin

import (
	"fmt"
	"regexp"
	"strings"
)

// validRegionPattern is the SSRF boundary for region identifiers used in OCI API calls.
// The character class [a-z0-9-] forbids '.', '/', ':', '@', whitespace, and every byte
// that could redirect endpoint construction — so any string this regex accepts is safe
// to interpolate into a host name. We intentionally do NOT maintain an allowlist of
// known regions: a static allowlist is a semantic check ("is this a known region?"),
// while the threat model needs a syntactic one ("is this a safe identifier?"). New OCI
// regions launch periodically, so an allowlist is also a recurring source of drift.
var validRegionPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}[a-z0-9]$`)

// validDomainPattern matches legitimate OCI domain components (e.g., "oraclecloud.com").
// Only allows lowercase alphanumeric, hyphens, and dots in FQDN format.
var validDomainPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)

// ValidateRegion checks whether a region string is safe to use in OCI API calls.
//
// Callers are expected to canonicalize input (TrimSpace + ToLower) at the boundary;
// the trim/lower below is a defence-in-depth no-op rather than a contract.
func ValidateRegion(region string) error {
	if region == "" {
		return fmt.Errorf("region cannot be empty")
	}

	region = strings.ToLower(strings.TrimSpace(region))

	if !validRegionPattern.MatchString(region) {
		return fmt.Errorf("invalid region identifier %q: must contain only lowercase alphanumeric characters and hyphens", region)
	}

	return nil
}

// ValidateCustomDomain checks whether a custom domain string is safe to use for endpoint construction.
// Only allows FQDN-format domains that look like legitimate OCI sovereign cloud domains.
func ValidateCustomDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("custom domain cannot be empty")
	}

	domain = strings.ToLower(strings.TrimSpace(domain))

	if len(domain) > 253 {
		return fmt.Errorf("custom domain %q exceeds maximum length", domain)
	}

	if !validDomainPattern.MatchString(domain) {
		return fmt.Errorf("invalid custom domain %q: must be a valid FQDN with only alphanumeric characters, hyphens, and dots", domain)
	}

	return nil
}

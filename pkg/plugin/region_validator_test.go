// Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package plugin

import (
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestValidateRegion_KnownRegions(t *testing.T) {
	validRegions := []string{
		"us-phoenix-1",
		"us-ashburn-1",
		"eu-frankfurt-1",
		"ap-sydney-1",
		"ap-tokyo-1",
		"uk-london-1",
		"sa-saopaulo-1",
		"me-jeddah-1",
		"ca-toronto-1",
	}

	for _, region := range validRegions {
		if err := ValidateRegion(region); err != nil {
			t.Errorf("ValidateRegion(%q) returned error for known region: %v", region, err)
		}
	}
}

func TestValidateRegion_CustomRegions(t *testing.T) {
	// Custom sovereign cloud regions should be accepted if they match the safe pattern
	validCustom := []string{
		"oracle-zurich-1",
		"custom-region-42",
		"eu-sovereign-1",
	}

	for _, region := range validCustom {
		if err := ValidateRegion(region); err != nil {
			t.Errorf("ValidateRegion(%q) rejected valid custom region: %v", region, err)
		}
	}
}

func TestValidateRegion_SSRFPayloads(t *testing.T) {
	maliciousRegions := []string{
		"",                               // empty
		"evil.com",                        // domain with dot
		"attacker.com/steal",              // path traversal with dot
		"http://evil.com",                 // URL scheme
		"https://evil.com",                // URL scheme
		"../../etc/passwd",                // path traversal
		"../../../attacker.com/metrics",   // path traversal
		"$(whoami)",                        // command injection
		"us-phoenix-1.evil.com",           // subdomain injection
		"telemetry.evil.com",              // service impersonation
		"169.254.169.254",                 // IMDS IP
		"localhost:8080",                  // localhost with port
		"us-phoenix-1\nevil.com",          // newline injection
		"us-phoenix-1\revil.com",          // carriage return injection
		" ",                               // whitespace only
		"a]b",                             // bracket injection
		"us-phoenix-1@evil.com",           // at-sign injection
	}

	for _, region := range maliciousRegions {
		if err := ValidateRegion(region); err == nil {
			t.Errorf("ValidateRegion(%q) should have rejected SSRF payload but accepted it", region)
		}
	}
}

func TestValidateRegion_EmptyString(t *testing.T) {
	if err := ValidateRegion(""); err == nil {
		t.Error("ValidateRegion(\"\") should return error for empty region")
	}
}

func TestValidateCustomDomain_ValidDomains(t *testing.T) {
	validDomains := []string{
		"oraclecloud.com",
		"oraclecloud.ch",
		"oci-cloud.example.com",
		"my-sovereign-cloud.oracle.co.uk",
	}

	for _, domain := range validDomains {
		if err := ValidateCustomDomain(domain); err != nil {
			t.Errorf("ValidateCustomDomain(%q) rejected valid domain: %v", domain, err)
		}
	}
}

func TestValidateCustomDomain_InvalidDomains(t *testing.T) {
	invalidDomains := []string{
		"",                          // empty
		"evil",                      // no TLD
		"../../../etc/passwd",       // path traversal
		"http://evil.com",           // URL scheme
		"evil.com/path",             // path component
		"evil.com:8080",             // port
		"-invalid.com",              // leading hyphen
		"invalid-.com",              // trailing hyphen in label
	}

	for _, domain := range invalidDomains {
		if err := ValidateCustomDomain(domain); err == nil {
			t.Errorf("ValidateCustomDomain(%q) should have rejected invalid domain but accepted it", domain)
		}
	}
}

func TestValidateCustomDomain_Empty(t *testing.T) {
	if err := ValidateCustomDomain(""); err == nil {
		t.Error("ValidateCustomDomain(\"\") should return error for empty domain")
	}
}

func TestOCILoadSettings_CredentialsFromSecureOnly(t *testing.T) {
	// This test verifies that credentials in plaintext JSONData are NOT loaded,
	// and only DecryptedSecureJSONData is used for sensitive fields.

	// Simulate plaintext JSONData containing a credential that should NOT be loaded
	jsonData := []byte(`{
		"profile0": "DEFAULT",
		"region0": "us-ashburn-1",
		"tenancy0": "SHOULD_NOT_BE_LOADED",
		"privkey0": "SHOULD_NOT_BE_LOADED",
		"user0": "SHOULD_NOT_BE_LOADED",
		"fingerprint0": "SHOULD_NOT_BE_LOADED"
	}`)

	// Simulate DecryptedSecureJSONData with the real credentials
	secureData := map[string]string{
		"tenancy0":     "ocid1.tenancy.oc1..real-tenancy",
		"user0":        "ocid1.user.oc1..real-user",
		"privkey0":     "-----BEGIN RSA PRIVATE KEY-----\nreal-key\n-----END RSA PRIVATE KEY-----",
		"fingerprint0": "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99",
	}

	req := mockDataSourceInstanceSettings(jsonData, secureData)
	config, err := OCILoadSettings(req)
	if err != nil {
		t.Fatalf("OCILoadSettings returned error: %v", err)
	}

	// Verify credentials come from secure data, not plaintext
	if config.tenancyocid["DEFAULT"] == "SHOULD_NOT_BE_LOADED" {
		t.Error("tenancy OCID was loaded from plaintext JSONData instead of DecryptedSecureJSONData")
	}
	if config.privkey["DEFAULT"] == "SHOULD_NOT_BE_LOADED" {
		t.Error("private key was loaded from plaintext JSONData instead of DecryptedSecureJSONData")
	}
	if config.user["DEFAULT"] == "SHOULD_NOT_BE_LOADED" {
		t.Error("user OCID was loaded from plaintext JSONData instead of DecryptedSecureJSONData")
	}
	if config.fingerprint["DEFAULT"] == "SHOULD_NOT_BE_LOADED" {
		t.Error("fingerprint was loaded from plaintext JSONData instead of DecryptedSecureJSONData")
	}

	// Verify the secure data was actually loaded
	if config.tenancyocid["DEFAULT"] != "ocid1.tenancy.oc1..real-tenancy" {
		t.Errorf("tenancy OCID = %q, want %q", config.tenancyocid["DEFAULT"], "ocid1.tenancy.oc1..real-tenancy")
	}
	if config.user["DEFAULT"] != "ocid1.user.oc1..real-user" {
		t.Errorf("user OCID = %q, want %q", config.user["DEFAULT"], "ocid1.user.oc1..real-user")
	}
}

func TestOCILoadSettings_BackwardCompatFallback(t *testing.T) {
	// When DecryptedSecureJSONData is empty (legacy deployment), credentials
	// should fall back to reading from plaintext JSONData for backward compatibility.
	jsonData := []byte(`{
		"profile0": "LEGACY",
		"region0": "us-ashburn-1",
		"tenancy0": "ocid1.tenancy.oc1..legacy-tenancy",
		"user0": "ocid1.user.oc1..legacy-user",
		"privkey0": "-----BEGIN RSA PRIVATE KEY-----\nlegacy-key\n-----END RSA PRIVATE KEY-----",
		"fingerprint0": "aa:bb:cc:dd"
	}`)

	// Empty secure data simulates a legacy deployment
	secureData := map[string]string{}

	req := mockDataSourceInstanceSettings(jsonData, secureData)
	config, err := OCILoadSettings(req)
	if err != nil {
		t.Fatalf("OCILoadSettings returned error for legacy config: %v", err)
	}

	// Legacy credentials should still be loaded for backward compatibility
	if config.tenancyocid["LEGACY"] != "ocid1.tenancy.oc1..legacy-tenancy" {
		t.Errorf("backward compat: tenancy = %q, want %q", config.tenancyocid["LEGACY"], "ocid1.tenancy.oc1..legacy-tenancy")
	}
	if config.user["LEGACY"] != "ocid1.user.oc1..legacy-user" {
		t.Errorf("backward compat: user = %q, want %q", config.user["LEGACY"], "ocid1.user.oc1..legacy-user")
	}
}

func TestValidateRegion_AllSubscribedRegionPassesRegex(t *testing.T) {
	// "all-subscribed-region" is a sentinel string, not a real region. Callers
	// (GetMonitoringClientForRegion, validateRegionParam) are responsible for
	// short-circuiting it via constants.ALL_REGION before reaching the validator;
	// the validator's job is purely syntactic. The sentinel happens to satisfy the
	// regex by design (lowercase letters and hyphens, length within bounds), and
	// this test pins that contract so a future refactor cannot silently break it.
	if err := ValidateRegion("all-subscribed-region"); err != nil {
		t.Errorf("expected ValidateRegion to accept the sentinel string syntactically, got %v", err)
	}
}

// TestOCILoadSettings_AllSixProfiles is the regression test for PR #357 (commit
// 4464475, "Fix OCILoadSettings DoS: infinite loop when all 6 profiles are set").
// The pre-fix code used reflection-based field iteration that looped indefinitely
// when every profile slot was populated. We assert two things:
//  1. The call returns within a small time budget (defends the loop bound).
//  2. All six profile keys are present in the parsed config.
func TestOCILoadSettings_AllSixProfiles(t *testing.T) {
	jsonData := []byte(`{
		"profile0": "P0", "region0": "us-phoenix-1",  "customregion0": "",
		"profile1": "P1", "region1": "us-ashburn-1",  "customregion1": "",
		"profile2": "P2", "region2": "eu-frankfurt-1","customregion2": "",
		"profile3": "P3", "region3": "ap-tokyo-1",    "customregion3": "",
		"profile4": "P4", "region4": "ap-sydney-1",   "customregion4": "",
		"profile5": "P5", "region5": "uk-london-1",   "customregion5": ""
	}`)

	secureData := map[string]string{
		"tenancy0": "ocid.t.0", "user0": "ocid.u.0", "privkey0": "k0", "fingerprint0": "f0",
		"tenancy1": "ocid.t.1", "user1": "ocid.u.1", "privkey1": "k1", "fingerprint1": "f1",
		"tenancy2": "ocid.t.2", "user2": "ocid.u.2", "privkey2": "k2", "fingerprint2": "f2",
		"tenancy3": "ocid.t.3", "user3": "ocid.u.3", "privkey3": "k3", "fingerprint3": "f3",
		"tenancy4": "ocid.t.4", "user4": "ocid.u.4", "privkey4": "k4", "fingerprint4": "f4",
		"tenancy5": "ocid.t.5", "user5": "ocid.u.5", "privkey5": "k5", "fingerprint5": "f5",
	}

	req := mockDataSourceInstanceSettings(jsonData, secureData)

	done := make(chan *OCIConfigFile, 1)
	errc := make(chan error, 1)
	go func() {
		c, err := OCILoadSettings(req)
		if err != nil {
			errc <- err
			return
		}
		done <- c
	}()

	select {
	case err := <-errc:
		t.Fatalf("OCILoadSettings returned error: %v", err)
	case config := <-done:
		for _, key := range []string{"P0", "P1", "P2", "P3", "P4", "P5"} {
			if _, ok := config.tenancyocid[key]; !ok {
				t.Errorf("profile %q missing from parsed config", key)
			}
		}
		if config.region["P0"] != "us-phoenix-1" || config.region["P5"] != "uk-london-1" {
			t.Errorf("region map not populated correctly: P0=%q P5=%q",
				config.region["P0"], config.region["P5"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("OCILoadSettings did not return within 2s with 6 profiles set; possible regression of PR #357 DoS fix")
	}
}

// TestOCILoadSettings_StopsAtFirstEmptyProfile pins the "break on first empty Profile"
// contract that the post-#357 implementation establishes — this is exactly the corner
// the old reflection-based logic mishandled.
func TestOCILoadSettings_StopsAtFirstEmptyProfile(t *testing.T) {
	jsonData := []byte(`{
		"profile0": "P0", "region0": "us-phoenix-1",
		"profile1": "P1", "region1": "us-ashburn-1",
		"profile2": "",   "region2": "",
		"profile3": "P3", "region3": "ap-tokyo-1"
	}`)
	secureData := map[string]string{
		"tenancy0": "t0", "user0": "u0", "privkey0": "k0", "fingerprint0": "f0",
		"tenancy1": "t1", "user1": "u1", "privkey1": "k1", "fingerprint1": "f1",
		"tenancy3": "t3", "user3": "u3", "privkey3": "k3", "fingerprint3": "f3",
	}
	req := mockDataSourceInstanceSettings(jsonData, secureData)

	config, err := OCILoadSettings(req)
	if err != nil {
		t.Fatalf("OCILoadSettings returned error: %v", err)
	}

	if _, ok := config.tenancyocid["P0"]; !ok {
		t.Error("expected P0 to be loaded")
	}
	if _, ok := config.tenancyocid["P1"]; !ok {
		t.Error("expected P1 to be loaded")
	}
	if _, ok := config.tenancyocid["P3"]; ok {
		t.Error("expected P3 to be skipped because P2 is empty (break-on-first-empty contract)")
	}
}

// mockDataSourceInstanceSettings creates a minimal backend.DataSourceInstanceSettings for testing.
func mockDataSourceInstanceSettings(jsonData []byte, secureData map[string]string) backend.DataSourceInstanceSettings {
	return backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: secureData,
	}
}

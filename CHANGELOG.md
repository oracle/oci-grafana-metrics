# Changelog

## [6.5.6](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.5.6) - 2026-07-08

### Fixed

- Fixed compartment pagination so the initial OCI request omits the page token and subsequent requests use the token returned by OCI.
- Added one bounded retry when OCI rejects a pagination token, without caching partial compartment results.

## [6.5.5](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.5.5) - 2026-06-24

### Security

- Security fix: upgraded `go.opentelemetry.io/otel/sdk` from `v1.40.0` to `v1.43.0` to remediate CVE-2026-39883.
- Security fix: removed unused `@babel/preset-env` dependency to eliminate the vulnerable `@babel/plugin-transform-modules-systemjs` path flagged for CVE-2026-44728.
- Security fix: bumped `fast-uri` to `4.0.0` to remediate CVE-2026-6321 and CVE-2026-6322.
- Security fix: bumped `form-data` to `4.0.6` to remediate CVE-2026-12143.
- Security fix: bumped `ws` to `8.21.0` to remediate CVE-2026-48779.
- Security fix: upgraded `github.com/grafana/grafana-plugin-sdk-go` from `v0.290.1` to `v0.292.1`.
- Security fix: upgraded `golang.org/x/crypto` to `v0.52.0` to remediate govulncheck findings GO-2026-5005, GO-2026-5006, GO-2026-5013, GO-2026-5014, GO-2026-5015, GO-2026-5016, GO-2026-5017, GO-2026-5018, GO-2026-5019, GO-2026-5020, GO-2026-5021, GO-2026-5023, and GO-2026-5033.
- Security fix: added yarn resolutions for additional OSV/Dependabot npm findings: `@babel/core` `7.29.7`, `@opentelemetry/core` `2.8.0`, `@protobufjs/utf8` `1.1.1`, `@tootallnate/once` `2.0.1`, `diff` `5.2.2`, `dompurify` `3.4.11`, `immutable` `4.3.8`, `js-cookie` `3.0.7`, `js-yaml` `4.2.0`, `lodash` `4.18.1`, `nanoid` `3.3.8`, `postcss` `8.5.10`, `prismjs` `1.30.0`, `protobufjs` `7.6.3`, `protocol-buffers-schema` `3.6.1`, `qs` `6.15.2`, `uplot` `1.6.31`, and `uuid` `11.1.1`.
- Tooling: bumped `typescript` to `5.2.2` and fixed the `ConfigEditor` props type so `yarn typecheck` passes after the dependency updates.

## [6.5.4](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.5.4) - 2026-05-05

### Security

- Security fix: prevent credential exposure and SSRF via the region parameter
- Security fix: resolve infinite loop DoS in OCILoadSettings when all 6 profiles are configured
- Security fix: upgrade picomatch and serialize-javascript to address new advisories

### Changed

- Backend logging: use correct levels and structured key-value format

## [6.5.3](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.5.3) - 2026-04-01

### Fixed

- Bug fix for issue #328: Explorer queries now use the selected region instead of always using the default/home region
- Fixed ALL_REGION aggregation to correctly query all subscribed regions

### Added

- Added per-region client pool for thread-safe, concurrent multi-region support
- Added sovereign cloud region discovery via AddRegionSchemaForPlc for DRCC/Alloy support

### Security

- Upgraded OCI Go SDK from v65.81.3 to v65.105.0
- Upgraded grafana-plugin-sdk-go from v0.250.0 to v0.290.1 to resolve CVE-2026-33186 (gRPC authorization bypass) and CVE-2026-24051 (OpenTelemetry SDK path hijacking)
- Bumped google.golang.org/grpc to v1.79.3 (CVE-2026-33186 fix)
- Added yarn resolutions for JS dev dependency vulnerabilities (form-data, flatted, minimatch, serialize-javascript)
- Added dev environment for multi-version Grafana plugin testing (v7.5, v9, v10, v11, v12)

## [Security Maintenance release (v6.5.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.5.2) - 28 Feb 2025

This release includes:

- security patches

## [Security Maintenance release (v6.5.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.5.1) - 28 Feb 2025

This release includes:

- Bug fix for issue #311 
- Bug fix for issue #309

## [Support for DRCC/Alloy Regions (v6.5.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.5.0) - 05 Feb 2025

This release includes:

- Support for DRCC/Alloy Regions
- Updated version of OCI Golang libraries to v65.81.3
- Some bug fixes
- Support for Custom Region

## [Security Maintenance release (v6.0.3)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.0.3) - 26 Nov 2024

This release includes:

- Implement a retry function for SummarizeMetrics operations

## [Security Maintenance release (v6.0.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.0.2) - 14 Oct 2024

This release includes:

- security patches
- Added new regions

## [Security Maintenance release (v6.0.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.0.1) - 03 Oct 2024

This release includes:

- security patches

## [Security Maintenance release (v6.0.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v6.0.0) - 16 Sep 2024

This release includes:

- security patches
- React 18 upgrades
- Support for Grafana 11
- new regions added
- golang and typescript libraries updates

## [Security Maintenance release (v5.5.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.5.1) - 03 Jul 2024

This release includes:

- few security patches
- bug fix in multiple tenancy test function

## [Alert enabled and Cross Tenancy support (v5.5.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.5.0) - 15 May 2024

This release includes:

- Alert support
- Cross Tenancy Support in instance principal mode
- bug fixes
- security fixes

## [Beta release which includes alerting support (v5.5.0-beta-unsigned)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.5.0-beta-unsigned) - 24 Apr 2024

************* WARNING ***********

This is a BETA release unsigned. It means is still not final, not ready for production and not signed. You must enable following option in grafana.ini in order to use this version:
app_mode = development
This version is NOT available in the Grafana catalogue yet, you need to manually install the plugin unzipping the binaries in your grafana plugin directory.

***********************************

This beta includes alerting function. 

## [Multi region support and interval enhanced (v5.2.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.2.0) - 16 Apr 2024

This release includes:
- Multi Region support
- use of interval as template var
- Auto interval setting
- security fix in package babel-traverse, vulnerable to CVE-2023-45133
- enhanced error handling 
- improved performances on queries

## [Regex fix and Sovereign cloud support (v5.1.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.1.1) - 08 Mar 2024

Implements the following:

- Regex fix when using Compartments in Template vars
- Sovereign Cloud Support
- Explore issue when using Grafana version above v10.1

## [Compartment regex fix and added new Sovereign regions (v5.1.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.1.0) - 07 Mar 2024

- Compartment regex fix.
-  Added new Sovereign regions

## [Added new regions for Data-source configuration (v5.0.4)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.4) - 20 Feb 2024

- Added new missing regions in Datasource Configuration
- Fixed a small bug  in data source configuration page(src/ConfigEditor.tsx)

## [Improved custom label management (v5.0.3)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.3) - 18 Jan 2024

This maintenance release includes:

Fix list of dimensions returned values from oci API in case of raw queries
Fix sorting bug in case custom labels are used for non-indexed dimensions (for example for oci_autonomous_database)

## [Raw query template vars (v5.0.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.2) - 10 Jan 2024

Raw query template vars

## [Raw Mode and Custom metrics labeling (v5.0.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.1) - 09 Jan 2024

- Support for labeling on some custom metrics
- Support for Raw query mode
- Bug fixes and security fixes

## [Grafana 10 support (v5.0.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.0) - 11 Oct 2023

- FE completely rewritten in React/Typescript
- Compatibility with Grafana 10
- Caching of region, tenancy, compartments, dimensions queries
- many performance improvements
- new Grafana API

## [Upgrade nodejs dependencies and fix minor security bugs (v4.0.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v4.0.1) - 11 Apr 2023

- Upgraded nodejs dependencies
- Upgraded net golang libraries

## [Multi-tenancy support and Secure JSON for OCI (v4.0.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v4.0.0) - 07 Mar 2023

This release features

- Multi-tenancy support
- Secure JSON secrets for OCI Configuration
- Added support for San Jose region

## [Region list sort, Customization of labels, ARM64 support, Namespace drop-down field bug fix and vulnerability patches (v3.0.6)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.6) - 24 Oct 2022

- Customization of Graph labels (https://github.com/oracle/oci-grafana-metrics/pull/131)
- OCI region list sorted in alphabetical order (https://github.com/oracle/oci-grafana-metrics/pull/132)
- Updated GoLang and Javascript libraries, vulnerability patches (https://github.com/oracle/oci-grafana-metrics/pull/133, https://github.com/oracle/oci-grafana-metrics/pull/134, https://github.com/oracle/oci-grafana-metrics/pull/136, https://github.com/oracle/oci-grafana-metrics/pull/137, https://github.com/oracle/oci-grafana-metrics/pull/138, https://github.com/oracle/oci-grafana-metrics/pull/139, https://github.com/oracle/oci-grafana-metrics/pull/140)
- Metrics Namespace drop-down field bug fix (https://github.com/oracle/oci-grafana-metrics/pull/135)
- ARM64 support (https://github.com/oracle/oci-grafana-metrics/pull/129)

## [v3.0.5 - Added new regions](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.5) - 16 Jun 2022

New regions added:
- Singapore
- Paris
- Queretaro

Update Grunt version.
Update Linux readme.

## [Minor patch (v3.0.4)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.4) - 16 Mar 2022

Update build files, plugin.json

## [Bugfix for template variables (v3.0.3)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.3) - 11 Feb 2022

Fixes issue with multiple data sources and template variables.

## [Grafana 8 hotfix support (v3.0.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.2) - 10 Feb 2022

Stopgap solution for Grafana 8 support which may not be fully polished. Full support will be released later.

## [UK Gov support added and dev changes (v2.2.4)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.2.4) - 09 Aug 2021

- Added  support for uk gov regions `uk-gov-london-1 , uk-gov-cardiff-1 `

Dev changes :
- Removed  toml file and replaced it with mod 
- Added some for installation and signing. 

For oracle grafana developers
In the next release : 
Append the following to the build script 
zip -r oci-grafana-metrics-<VERSION> ./dist

## [Missing metrics fixed (v2.2.3)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.2.3) - 22 Jan 2021

In test phase, please don't use in  production

- Now,  all the metrics are being received. 
- The metrics request is 20x faster now. 
- Signed the plugin and added to tar file. 

## [Fixed documentation w.r.t resource group (v2.2.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.2.2) - 11 Jan 2021

- Fixed documentation  with resource group in each query 

## [Added support for dubai, santiago and cadiff (v.2.2.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v.2.2.1) - 22 Dec 2020

This release adds support to new regions such as santiago,   cardiff and dubai

## [ap-chiyoda-1  support   added (v2.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.2) - 08 Dec 2020

- Supports calling Oracle Cloud Infrastructure services in the ap-chiyoda-1 region 

## [Metrics plugin for grafana support (2.0.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/2.0.2) - 05 Oct 2020

- Please download only the plugin file 

- This is not for all users

- Please generate the manifest file and send  us back in email 

- The version has been updated to 2.0.0 in plugin.json

## [Logging-plugin-zip (v2.0.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.0.1) - 05 Oct 2020

- Not for all all users 
- This is not associated with the current code 
- Download only plugin.tar and other 
- Only for grafana support 

## [v2.0.0](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.0.0) - 28 Sep 2020

- Updated name and id of the plugin. 

- Please remove the existing plugin and install this version

- There is no functional change. 

## [Support Grafana 7 (v1.1.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.1.2) - 07 Jul 2020

- Update docs for Grafana 7
- Expand the default region list.
- Fix resolution Input

## [List metrics  api call fix (v1.1.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.1.1) - 04 Jun 2020

- Now, the number of  list metrics call has been  set to a maximum of 20  pages 
- The limit is configurable

## [Support added for variables and auto in window & resolution (v1.1.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.1.0) - 06 Apr 2020

New Features
- Support added for variables and auto in window & resolution
###### Minor bug fixes
- Fixed variable duplication in dropdowns 
###### Documentation
- Instructions added for using variables and auto-config in window and resolution of drop down

## [Added a region option to getCompartment method (V1.0.9)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.9) - 12 Mar 2020

When OCI tenancy is provision with a single region, say 'us-phoenix-1' (home tenancy) and the datasource setting with local has the default regions as 'us-ausburn-1' (in the ~/.oci/config file), the getCompartment does not return the compartment list. The reason being the OCI tenancy has only one tenancy as home and that is not the 'us-phoenix-1' tenancy.

The fix is to set region to the home regions of the OCI tenancy, while making the getCompartment call.

Thanks Jayesh Patel

## [Support Resource Group (V1.0.8)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.8) - 10 Mar 2020

User will be able to use Resource Group.
By Default: No Resource Group

## [V1.0.7](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.7) - 21 Feb 2020

Fix populated Metric Fields based on the selected region.

## [Support hard coded values in template variables (V1.0.6)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.6) - 23 Jan 2020

1. Support hard coded values in template variables
2. Fix bug with filtering compartments by regex in template variable editor

## [Manual query support. New template variables for dimensions. (V1.0.5)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.5) - 25 Nov 2019

1. New template variables were added: 
 - `dimensions()` which show all possible dimension keys for selected region, compartment, namespace and metric
- `dimensionOptions()` which show all possible dimension values for selected region, compartment, namespace, metric and dimension key.

2. Dimension value variable can be used as multi-value. Separate queries are generated for each dimension value out of multi-select (but no more than 20 queries)

3. Custom query is supported. User can type MQL expression manually which will be passed to telemetry instead of selected on the UI metric and dimensions.

4. Custom template variables were added to the list of dimension value options. Duplicate options are removed.

5. Options for regions and compartments are cached while query editor is open. Dimension options are cached for the selected region - compartment - namespace - metric.

## [Fix $namespace and $metric variables (V1.0.4)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.4) - 03 Oct 2019

`$namespace`
In the previous version, the list of namespaces was hardcoded for the $namespace variable. 
Today the list of namespaces depends on region and compartment.

`$metric`
In the previous version, the $metric variable depends on the home region, $compartment and hardcoded $namespace. 
Today $metric depends on $region,  $compartment and $namespace.

## [V1.0.3](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.3) - 25 Sep 2019

Fix issue: 
Metric name rule creation updated

## [v1.0.2](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.0.2) - 27 Mar 2019

- Shows subcompartments and removes inactive compartments from that list
- Changes Metric names when rendering on the screen, uses a human readable name for the resource if one is present
- Shortens ocids to first three and last six characters to save screen real estate 

## [v1.0.1](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.0.1) - 08 Mar 2019

Pulls in regions dynamically
Adds more documentation

<!-- Generated by https://github.com/rhysd/changelog-from-release v3.7.2 -->

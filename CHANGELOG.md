<a name="v5.5.1"></a>
# [Security Maintenance release (v5.5.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.5.1) - 03 Jul 2024

This release includes:

- few security patches
- bug fix in multiple tenancy test function

[Changes][v5.5.1]


<a name="v5.5.0"></a>
# [Alert enabled and Cross Tenancy support (v5.5.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.5.0) - 15 May 2024

This release includes:

- Alert support
- Cross Tenancy Support in instance principal mode
- bug fixes
- security fixes

[Changes][v5.5.0]


<a name="v5.5.0-beta-unsigned"></a>
# [Beta release which includes alerting support (v5.5.0-beta-unsigned)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.5.0-beta-unsigned) - 24 Apr 2024

************* WARNING ***********

This is a BETA release unsigned. It means is still not final, not ready for production and not signed. You must enable following option in grafana.ini in order to use this version:
app_mode = development
This version is NOT available in the Grafana catalogue yet, you need to manually install the plugin unzipping the binaries in your grafana plugin directory.

***********************************

This beta includes alerting function. 

[Changes][v5.5.0-beta-unsigned]


<a name="v5.2.0"></a>
# [Multi region support and interval enhanced (v5.2.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.2.0) - 16 Apr 2024

This release includes:
- Multi Region support
- use of interval as template var
- Auto interval setting
- security fix in package babel-traverse, vulnerable to CVE-2023-45133
- enhanced error handling 
- improved performances on queries

[Changes][v5.2.0]


<a name="v5.1.1"></a>
# [Regex fix and Sovereign cloud support (v5.1.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.1.1) - 08 Mar 2024

Implements the following:

- Regex fix when using Compartments in Template vars
- Sovereign Cloud Support
- Explore issue when using Grafana version above v10.1

[Changes][v5.1.1]


<a name="v5.1.0"></a>
# [Compartment regex fix and added new Sovereign regions (v5.1.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.1.0) - 07 Mar 2024

- Compartment regex fix.
-  Added new Sovereign regions

[Changes][v5.1.0]


<a name="v5.0.4"></a>
# [Added new regions for Data-source configuration (v5.0.4)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.4) - 20 Feb 2024

- Added new missing regions in Datasource Configuration
- Fixed a small bug  in data source configuration page(src/ConfigEditor.tsx)

[Changes][v5.0.4]


<a name="v5.0.3"></a>
# [Improved custom label management (v5.0.3)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.3) - 18 Jan 2024

This maintenance release includes:

Fix list of dimensions returned values from oci API in case of raw queries
Fix sorting bug in case custom labels are used for non-indexed dimensions (for example for oci_autonomous_database)

[Changes][v5.0.3]


<a name="v5.0.2"></a>
# [Raw query template vars (v5.0.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.2) - 10 Jan 2024

Raw query template vars

[Changes][v5.0.2]


<a name="v5.0.1"></a>
# [Raw Mode and Custom metrics labeling (v5.0.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.1) - 09 Jan 2024

- Support for labeling on some custom metrics
- Support for Raw query mode
- Bug fixes and security fixes

[Changes][v5.0.1]


<a name="v5.0.0"></a>
# [Grafana 10 support (v5.0.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v5.0.0) - 11 Oct 2023

- FE completely rewritten in React/Typescript
- Compatibility with Grafana 10
- Caching of region, tenancy, compartments, dimensions queries
- many performance improvements
- new Grafana API

[Changes][v5.0.0]


<a name="v4.0.1"></a>
# [Upgrade nodejs dependencies and fix minor security bugs (v4.0.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v4.0.1) - 11 Apr 2023

- Upgraded nodejs dependencies
- Upgraded net golang libraries

[Changes][v4.0.1]


<a name="v4.0.0"></a>
# [Multi-tenancy support and Secure JSON for OCI (v4.0.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v4.0.0) - 07 Mar 2023

This release features

- Multi-tenancy support
- Secure JSON secrets for OCI Configuration
- Added support for San Jose region

[Changes][v4.0.0]


<a name="v3.0.6"></a>
# [Region list sort, Customization of labels, ARM64 support, Namespace drop-down field bug fix and vulnerability patches (v3.0.6)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.6) - 24 Oct 2022

- Customization of Graph labels (https://github.com/oracle/oci-grafana-metrics/pull/131)
- OCI region list sorted in alphabetical order (https://github.com/oracle/oci-grafana-metrics/pull/132)
- Updated GoLang and Javascript libraries, vulnerability patches (https://github.com/oracle/oci-grafana-metrics/pull/133, https://github.com/oracle/oci-grafana-metrics/pull/134, https://github.com/oracle/oci-grafana-metrics/pull/136, https://github.com/oracle/oci-grafana-metrics/pull/137, https://github.com/oracle/oci-grafana-metrics/pull/138, https://github.com/oracle/oci-grafana-metrics/pull/139, https://github.com/oracle/oci-grafana-metrics/pull/140)
- Metrics Namespace drop-down field bug fix (https://github.com/oracle/oci-grafana-metrics/pull/135)
- ARM64 support (https://github.com/oracle/oci-grafana-metrics/pull/129)

[Changes][v3.0.6]


<a name="v3.0.5"></a>
# [v3.0.5 - Added new regions](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.5) - 16 Jun 2022

New regions added:
- Singapore
- Paris
- Queretaro

Update Grunt version.
Update Linux readme.

[Changes][v3.0.5]


<a name="v3.0.4"></a>
# [Minor patch (v3.0.4)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.4) - 16 Mar 2022

Update build files, plugin.json

[Changes][v3.0.4]


<a name="v3.0.3"></a>
# [Bugfix for template variables (v3.0.3)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.3) - 11 Feb 2022

Fixes issue with multiple data sources and template variables.

[Changes][v3.0.3]


<a name="v3.0.2"></a>
# [Grafana 8 hotfix support (v3.0.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v3.0.2) - 10 Feb 2022

Stopgap solution for Grafana 8 support which may not be fully polished. Full support will be released later.

[Changes][v3.0.2]


<a name="v2.2.4"></a>
# [UK Gov support added and dev changes (v2.2.4)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.2.4) - 09 Aug 2021

- Added  support for uk gov regions `uk-gov-london-1 , uk-gov-cardiff-1 `

Dev changes :
- Removed  toml file and replaced it with mod 
- Added some for installation and signing. 

For oracle grafana developers
In the next release : 
Append the following to the build script 
zip -r oci-grafana-metrics-<VERSION> ./dist

[Changes][v2.2.4]


<a name="v2.2.3"></a>
# [Missing metrics fixed (v2.2.3)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.2.3) - 22 Jan 2021

In test phase, please don't use in  production

- Now,  all the metrics are being received. 
- The metrics request is 20x faster now. 
- Signed the plugin and added to tar file. 


[Changes][v2.2.3]


<a name="v2.2.2"></a>
# [Fixed documentation w.r.t resource group (v2.2.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.2.2) - 11 Jan 2021



- Fixed documentation  with resource group in each query 

[Changes][v2.2.2]


<a name="v.2.2.1"></a>
# [Added support for dubai, santiago and cadiff (v.2.2.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v.2.2.1) - 22 Dec 2020

This release adds support to new regions such as santiago,   cardiff and dubai

[Changes][v.2.2.1]


<a name="v2.2"></a>
# [ap-chiyoda-1  support   added (v2.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.2) - 08 Dec 2020

- Supports calling Oracle Cloud Infrastructure services in the ap-chiyoda-1 region 



[Changes][v2.2]


<a name="2.0.2"></a>
# [Metrics plugin for grafana support (2.0.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/2.0.2) - 05 Oct 2020

- Please download only the plugin file 

- This is not for all users

- Please generate the manifest file and send  us back in email 

- The version has been updated to 2.0.0 in plugin.json

[Changes][2.0.2]


<a name="v2.0.1"></a>
# [Logging-plugin-zip (v2.0.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.0.1) - 05 Oct 2020

- Not for all all users 
- This is not associated with the current code 
- Download only plugin.tar and other 
- Only for grafana support 

[Changes][v2.0.1]


<a name="v2.0.0"></a>
# [v2.0.0](https://github.com/oracle/oci-grafana-metrics/releases/tag/v2.0.0) - 28 Sep 2020

- Updated name and id of the plugin. 

- Please remove the existing plugin and install this version

- There is no functional change. 



[Changes][v2.0.0]


<a name="v1.1.2"></a>
# [Support Grafana 7 (v1.1.2)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.1.2) - 07 Jul 2020

- Update docs for Grafana 7
- Expand the default region list.
- Fix resolution Input

[Changes][v1.1.2]


<a name="v1.1.1"></a>
# [List metrics  api call fix (v1.1.1)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.1.1) - 04 Jun 2020

- Now, the number of  list metrics call has been  set to a maximum of 20  pages 
- The limit is configurable


[Changes][v1.1.1]


<a name="v1.1.0"></a>
# [Support added for variables and auto in window & resolution (v1.1.0)](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.1.0) - 06 Apr 2020

New Features
- Support added for variables and auto in window & resolution
###### Minor bug fixes
- Fixed variable duplication in dropdowns 
###### Documentation
- Instructions added for using variables and auto-config in window and resolution of drop down





[Changes][v1.1.0]


<a name="V1.0.9"></a>
# [Added a region option to getCompartment method (V1.0.9)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.9) - 12 Mar 2020

When OCI tenancy is provision with a single region, say 'us-phoenix-1' (home tenancy) and the datasource setting with local has the default regions as 'us-ausburn-1' (in the ~/.oci/config file), the getCompartment does not return the compartment list. The reason being the OCI tenancy has only one tenancy as home and that is not the 'us-phoenix-1' tenancy.

The fix is to set region to the home regions of the OCI tenancy, while making the getCompartment call.

Thanks Jayesh Patel

[Changes][V1.0.9]


<a name="V1.0.8"></a>
# [Support Resource Group (V1.0.8)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.8) - 10 Mar 2020

User will be able to use Resource Group.
By Default: No Resource Group

[Changes][V1.0.8]


<a name="V1.0.7"></a>
# [V1.0.7](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.7) - 21 Feb 2020

Fix populated Metric Fields based on the selected region.

[Changes][V1.0.7]


<a name="V1.0.6"></a>
# [Support hard coded values in template variables (V1.0.6)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.6) - 23 Jan 2020

1. Support hard coded values in template variables
2. Fix bug with filtering compartments by regex in template variable editor

[Changes][V1.0.6]


<a name="V1.0.5"></a>
# [Manual query support. New template variables for dimensions. (V1.0.5)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.5) - 25 Nov 2019

1. New template variables were added: 
 - `dimensions()` which show all possible dimension keys for selected region, compartment, namespace and metric
- `dimensionOptions()` which show all possible dimension values for selected region, compartment, namespace, metric and dimension key.

2. Dimension value variable can be used as multi-value. Separate queries are generated for each dimension value out of multi-select (but no more than 20 queries)

3. Custom query is supported. User can type MQL expression manually which will be passed to telemetry instead of selected on the UI metric and dimensions.

4. Custom template variables were added to the list of dimension value options. Duplicate options are removed.

5. Options for regions and compartments are cached while query editor is open. Dimension options are cached for the selected region - compartment - namespace - metric.

[Changes][V1.0.5]


<a name="V1.0.4"></a>
# [Fix $namespace and $metric variables (V1.0.4)](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.4) - 03 Oct 2019

`$namespace`
In the previous version, the list of namespaces was hardcoded for the $namespace variable. 
Today the list of namespaces depends on region and compartment.

`$metric`
In the previous version, the $metric variable depends on the home region, $compartment and hardcoded $namespace. 
Today $metric depends on $region,  $compartment and $namespace.



[Changes][V1.0.4]


<a name="V1.0.3"></a>
# [V1.0.3](https://github.com/oracle/oci-grafana-metrics/releases/tag/V1.0.3) - 25 Sep 2019

Fix issue: 
Metric name rule creation updated

[Changes][V1.0.3]


<a name="v1.0.2"></a>
# [v1.0.2](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.0.2) - 27 Mar 2019

- Shows subcompartments and removes inactive compartments from that list
- Changes Metric names when rendering on the screen, uses a human readable name for the resource if one is present
- Shortens ocids to first three and last six characters to save screen real estate 


[Changes][v1.0.2]


<a name="v1.0.1"></a>
# [v1.0.1](https://github.com/oracle/oci-grafana-metrics/releases/tag/v1.0.1) - 08 Mar 2019

Pulls in regions dynamically
Adds more documentation

[Changes][v1.0.1]


[v5.5.1]: https://github.com/oracle/oci-grafana-metrics/compare/v5.5.0...v5.5.1
[v5.5.0]: https://github.com/oracle/oci-grafana-metrics/compare/v5.5.0-beta-unsigned...v5.5.0
[v5.5.0-beta-unsigned]: https://github.com/oracle/oci-grafana-metrics/compare/v5.2.0...v5.5.0-beta-unsigned
[v5.2.0]: https://github.com/oracle/oci-grafana-metrics/compare/v5.1.1...v5.2.0
[v5.1.1]: https://github.com/oracle/oci-grafana-metrics/compare/v5.1.0...v5.1.1
[v5.1.0]: https://github.com/oracle/oci-grafana-metrics/compare/v5.0.4...v5.1.0
[v5.0.4]: https://github.com/oracle/oci-grafana-metrics/compare/v5.0.3...v5.0.4
[v5.0.3]: https://github.com/oracle/oci-grafana-metrics/compare/v5.0.2...v5.0.3
[v5.0.2]: https://github.com/oracle/oci-grafana-metrics/compare/v5.0.1...v5.0.2
[v5.0.1]: https://github.com/oracle/oci-grafana-metrics/compare/v5.0.0...v5.0.1
[v5.0.0]: https://github.com/oracle/oci-grafana-metrics/compare/v4.0.1...v5.0.0
[v4.0.1]: https://github.com/oracle/oci-grafana-metrics/compare/v4.0.0...v4.0.1
[v4.0.0]: https://github.com/oracle/oci-grafana-metrics/compare/v3.0.6...v4.0.0
[v3.0.6]: https://github.com/oracle/oci-grafana-metrics/compare/v3.0.5...v3.0.6
[v3.0.5]: https://github.com/oracle/oci-grafana-metrics/compare/v3.0.4...v3.0.5
[v3.0.4]: https://github.com/oracle/oci-grafana-metrics/compare/v3.0.3...v3.0.4
[v3.0.3]: https://github.com/oracle/oci-grafana-metrics/compare/v3.0.2...v3.0.3
[v3.0.2]: https://github.com/oracle/oci-grafana-metrics/compare/v2.2.4...v3.0.2
[v2.2.4]: https://github.com/oracle/oci-grafana-metrics/compare/v2.2.3...v2.2.4
[v2.2.3]: https://github.com/oracle/oci-grafana-metrics/compare/v2.2.2...v2.2.3
[v2.2.2]: https://github.com/oracle/oci-grafana-metrics/compare/v.2.2.1...v2.2.2
[v.2.2.1]: https://github.com/oracle/oci-grafana-metrics/compare/v2.2...v.2.2.1
[v2.2]: https://github.com/oracle/oci-grafana-metrics/compare/2.0.2...v2.2
[2.0.2]: https://github.com/oracle/oci-grafana-metrics/compare/v2.0.1...2.0.2
[v2.0.1]: https://github.com/oracle/oci-grafana-metrics/compare/v2.0.0...v2.0.1
[v2.0.0]: https://github.com/oracle/oci-grafana-metrics/compare/v1.1.2...v2.0.0
[v1.1.2]: https://github.com/oracle/oci-grafana-metrics/compare/v1.1.1...v1.1.2
[v1.1.1]: https://github.com/oracle/oci-grafana-metrics/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/oracle/oci-grafana-metrics/compare/V1.0.9...v1.1.0
[V1.0.9]: https://github.com/oracle/oci-grafana-metrics/compare/V1.0.8...V1.0.9
[V1.0.8]: https://github.com/oracle/oci-grafana-metrics/compare/V1.0.7...V1.0.8
[V1.0.7]: https://github.com/oracle/oci-grafana-metrics/compare/V1.0.6...V1.0.7
[V1.0.6]: https://github.com/oracle/oci-grafana-metrics/compare/V1.0.5...V1.0.6
[V1.0.5]: https://github.com/oracle/oci-grafana-metrics/compare/V1.0.4...V1.0.5
[V1.0.4]: https://github.com/oracle/oci-grafana-metrics/compare/V1.0.3...V1.0.4
[V1.0.3]: https://github.com/oracle/oci-grafana-metrics/compare/v1.0.2...V1.0.3
[v1.0.2]: https://github.com/oracle/oci-grafana-metrics/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/oracle/oci-grafana-metrics/tree/v1.0.1

<!-- Generated by https://github.com/rhysd/changelog-from-release v3.7.2 -->

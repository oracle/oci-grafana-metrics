# About OCI Metrics plugin for Grafana

  
## Introduction 
This plugin makes queries to the Oracle Cloud Infrastructure(OCI) Monitoring Service to fetch metrics for your OCI resources or your custom metrics in OCI. And then displays them on Grafana.


If you are running Grafana on a machine instance in Oracle Cloud, use the Instance Principal with a configured Dynamic Group and policy to allow you to read metrics and compartments.
  

If you are running Grafana anywhere else you'll need to get the necessary provider and resource settings, as described in this section [Getting OCI Configuration values](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/linux.md#getting-oci-configuration-values)
  

Latest plugin version 5.X.X(available on [Grafana Marketplace](https://grafana.com/grafana/plugins/oci-metrics-datasource/)) is compatible with **Grafana 10**. We will release it's binary on [its Github repo](https://github.com/oracle/oci-grafana-plugin) very soon.
  
Breaking changes:
- In case you are migrating from a previous version (3.x.x or below) of the OCI Metrics Grafana Plugin and are not using Instance Principals (Environment not set as OCI instance), please refer to the [**Migration Instructions for Grafana OCI Metrics Data Source Settings (User Principals and Single Tenancy mode only)**](https://github.com/oracle/oci-grafana-metrics/blob/master/docs/migration.md) because you will have to reconfigure the plugin setup
- In case you are migrating from a previous version (4.x.x or below) of the OCI Metrics Grafana Plugin to version 5.x.x and your dashboard is using **dimensions** in its selectors or as its template variables, please refer to [this section](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/using.md#migrate-to-version-5.x) to modify your dashboard accordingly to the new **dimensions** selector. 

Oracle Cloud Infrastructure Metrics plugin is *datasource with backend* type of plugin Grafana. Hereafter referred to as OCI Metrics plugin. 

## Installation
Please refers to the following **compatibility matrix** to choose pluiigin version accordingly to your Grafana installation: [Compatibility Matrix](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/compatmatrix.md)

In order to simplify the installation process, we created detailed guides for you to follow. 

* Install Grafana and the OCI Metrics plugin on a Linux host using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/linux.md).

* Install Grafana and the OCI Metrics plugin on Grafana Cloud using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/grafanacloud.md).

* Install Grafana and the OCI Metrics plugin on a MacOS host using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/macos.md).

* Install Grafana and the OCI Metrics plugin on a virtual machine in Oracle Cloud Infrastructure using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/linuxoci.md).

* Install Grafana and the OCI Metrics plugin on a virtual machine in Oracle Cloud Infrastructure using Terraform using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/terraform.md).

* Install Grafana and the OCI Metrics plugin on Kubernetes in Oracle Cloud Infrastructure using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/kubernetes.md)

  

Once you have the OCI Metrics Plugin installed, configure your datasource with your tenancy OCID, default region, and right IAM setup(Dynamic Group or OCI User Auth with local Private key file on Grafana Server node-depending where you're running the Grafana-Oracle Cloud or elsewhere).

  

We also have documentation for how to use the newly installed and configured plugin in our [Using Grafana with OCI Metric Plugin](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/using.md) walkthrough.

## Note 1

If you're using a version of Grafana that's older than 6.0, you will need to download the zip file for plugin versions <=2.2.4 and install this plugin manually, or chmod the binary that is downloaded to make it executable. 

## Note 2

The OCI Metrics plugin supports the integration with Grafana Cloud with Data Source **Environment** configured as **local**. See [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/grafanacloud.md) for additional information.

### Debugging

Please make sure that the golang version installed is ```1.16``` and grafana version installed is 8.x.x

If you want to debug golang backend plugin code, follow the steps below:

* Install [gops](https://github.com/google/gops) to list running go processes on your machine

* Run `gops` and find processId for `oci-plugin_darwin_amd64` process

* Copy this processId to the `.vscode/launch.json`

* In your VSCode from 'Debug' menu call 'Start Debugging'

  

## Documentation

  

Please refer to the [docs folder in this repo](https://github.com/oracle/oci-grafana-metrics/tree/master/docs)

  

## Help

  

Issues and questions about this plugin can be posted [as an issue in this GitHub repository](https://github.com/oracle/oci-grafana-plugin/issues)

  

## Contributing


This project welcomes contributions from the community. Before submitting a pull request, please [review our contribution guide](https://github.com/oracle/oci-grafana-metrics/blob/master/CONTRIBUTING.md)

  

## Security

  

Please consult the [security guide](https://github.com/oracle/oci-grafana-metrics/blob/master/SECURITY.md) for our responsible security

vulnerability disclosure process.

  

## License

  

Copyright (c) 2023 Oracle and/or its affiliates.

  

Released under the Universal Permissive License v1.0 as shown at

<https://oss.oracle.com/licenses/upl/>.


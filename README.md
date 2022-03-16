# About OCI Metrics plugin for Grafana

  
## Introduction 
This plugin makes queries to the Oracle Cloud Infrastructure(OCI) Monitoring Service to fetch metrics for your OCI resources or your custom metrics in OCI. And then displays them on Grafana.


If you are running Grafana on a machine instance in Oracle Cloud, use the Instance Principal with a configured Dynamic Group and policy to allow you to read metrics and compartments.
  

If you are running Grafana anywhere else, make sure you have `~/.oci` configured properly. You can do this by installing the Oracle Cloud CLI and running the setup.
  

Latest plugin version 3.X.X(available on [Grafana Marketplace](https://grafana.com/grafana/plugins/oci-metrics-datasource/)) is compatible with Grafana 8**. We will release it's binary on [its Github repo](https://github.com/oracle/oci-grafana-plugin) very soon.
  
Oracle Cloud Infrastructure Metrics plugin is *datasource with backend* type of plugin Grafana. Hereafter referred to as OCI Metrics plugin.  

## Installation
*We highly recommend to use Grafana 8.x.x with plugin version 3.x.x .*
In order to simplify the installation process, we created detailed guides for you to follow. 


* Install Grafana and the OCI Metrics plugin on a Linux host using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/linux.md).

* Install Grafana and the OCI Metrics plugin on a MacOS host using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/macos.md).

* Install Grafana and the OCI Metrics plugin on a virtual machine in Oracle Cloud Infrastructure using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/linuxoci.md).

* Install Grafana and the OCI Metrics plugin on a virtual machine in Oracle Cloud Infrastructure using Terraform using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/terraform.md).

* Install Grafana and the OCI Metrics plugin on Kubernetes in Oracle Cloud Infrastructure using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/kubernetes.md)

  

Once you have the OCI Metrics Plugin installed, configure your datasource with your tenancy OCID, default region, and right IAM setup(Dynamic Group or OCI User Auth with local Private key file on Grafana Server node-depending where you're running the Grafana-Oracle Cloud or elsewhere).

  

We also have documentation for how to use the newly installed and configured plugin in our [Using Grafana with OCI Metric Plugin](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/using.md) walkthrough.

## Note 1

If you're using a version of Grafana that's older than 6.0, you will need to download the zip file for plugin versions <=2.2.4 and install this plugin manually, or chmod the binary that is downloaded to make it executable. 


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

  

This project welcomes contributions from the community. Before submitting a pull

request, please [review our contribution guide](https://github.com/oracle/oci-grafana-metrics/blob/master/CONTRIBUTING.md).

  

## Security

  

Please consult the [security guide](https://github.com/oracle/oci-grafana-metrics/blob/master/SECURITY.md) for our responsible security

vulnerability disclosure process.

  

## License

  

Copyright (c) 2021 Oracle and/or its affiliates.

  

Released under the Universal Permissive License v1.0 as shown at

<https://oss.oracle.com/licenses/upl/>.


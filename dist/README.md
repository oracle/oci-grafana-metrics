# Oracle Cloud Infrastructure Data Source for Grafana

## Introduction

This plugin makes queries to the Oracle Cloud Infrastructure Telemetry Service and displays them on Grafana.

If you are running Grafana on a machine instance in Oracle Cloud, use the Service Principal with a configured Dynamic Group and policy to allow you to read metrics and compartments.

If you are running Grafana anywhere else, make sure you have `~/.oci` configured properly. You can do this by installing the Oracle Cloud CLI and running the setup 

## Note

If you're using a version of Grafana that's older 6.0, you will need to download the zip file and install this plugin manually, or chmod the binary that is downloaded to make it executable.

## Installation Documentation

In order to simplify the installation process, we created detailed guides for you to follow:

* Install Grafana and the Oracle Cloud Infrastructure Data Source for Grafana on a Linux host using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/Local%20Installation%20Linux%20-%20Oracle%20Cloud%20Infrastructure%20Data%20Source%20for%20Grafana.md).
* Install Grafana and the Oracle Cloud Infrastructure Data Source for Grafana on a MacOS host using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/Local%20Installation%20MacOS%20-%20Oracle%20Cloud%20Infrastructure%20Data%20Source%20for%20Grafana.md).
* Install Grafana and the Oracle Cloud Infrastructure Data Source for Grafana on a virtual machine in Oracle Cloud Infrastructure using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/OCI%20Virtual%20Machine%20Installation%20-%20Oracle%20Cloud%20Infrastructure%20Data%20Source%20for%20Grafana.md).
* Install Grafana and the Oracle Cloud Infrastructure Data Source for Grafana on a virtual machine in Oracle Cloud Infrastructure using Terraform using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/Terraform-Based%20Installation%20-%20Oracle%20Cloud%20Infrastructure%20Data%20Source%20for%20Grafana.md).
* Install Grafana and the Oracle Cloud Infrastructure Data Source for Grafana on Kubernetes in Oracle Cloud Infrastructure using [this document](https://github.com/oracle/oci-grafana-plugin/blob/master/docs/Oracle%20Kubernetes%20Engine%20Installation%20-%20Oracle%20Cloud%20Infrastructure%20Data%20Source%20for%20Grafana.md)

Once you have the data source installed, configure your datasource with your tenancy OCID, default region, and where you're running the plugin (Oracle Cloud or elsewhere).


# Terraform-Based Installation - Oracle Cloud Infrastructure Data Source for Grafana

## Background

Grafana is a popular technology that makes it easy to visualize metrics. The Oracle Cloud Infrastructure Data Source for Grafana is used to extend Grafana by adding OCI as a data source. The plugin enables you to visualize metrics related to a number of OCI resources: Compute, Networking, Storage, and custom metrics. 

This walkthrough is intended for use by people who would like to deploy Grafana and the OCI Data Source for Grafana on a virtual machine in OCI. 

## Install Terraform 

Begin by [installing Terraform](https://learn.hashicorp.com/terraform/getting-started/install) onto your development machine. We will be using the [Oracle Cloud Infrastructure Provider](https://www.terraform.io/docs/providers/oci/index.html), a tool designed to understand API interactions and exposing resource related to a specific cloud vendor, in this case Oracle Cloud Infrastructure. Terraform requires a number of variables in order to properly create an environment for Grafana and the OCI Data Source for Grafana plugin: 

```
variable "tenancy_ocid" {}
variable "user_ocid" {}
variable "fingerprint" {}
variable "private_key_path" {}
variable "region" {}
variable "compartment_ocid" {}
variable "ssh_public_key" {}
variable "ssh_private_key" {}
variable "subnet_id" {}
variable "availability_domain" {}
variable "dynamic_group_name" {}
```

Information regarding how to find each of these variables is located [here](https://docs.cloud.oracle.com/iaas/Content/API/Concepts/apisigningkey.htm). It can be useful to pass these variables into your .bashrc or .bash_profile. You will need to enter a value for the availability domain of your subnet (e.g. PKGK:US-ASHBURN-AD-1). This information can be found under Networking > Virtual Cloud Networks > [Your Network] > Subnets. It will also prompt you for the name of your dynamic group (e.g. metric-collection). 

The Terraform script requires the use of an existing [Virtual Cloud Network](https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingVCNs.htm) with access to the public internet. If you do not already have access to a Virtual Cloud Network with access to the public internet you can navigate to **Virtual Cloud Networks** under **Networking** and click **Create Virtual
Cloud Network**. Choosing the **CREATE VIRTUAL CLOUD NETWORK PLUS RELATED RESOURCES** option will result in a VCN with an Internet Routing Gateway and Route Tables configured for access to the public internet. Three subnets will be created: one in each availability domain in the region. 

![Screen Shot 2018-12-17 at 3.58.23 PM](images/Screen%20Shot%202018-12-17%20at%203.58.23%20PM.png)

## Create the Grafana Environment with Terraform 

After Terraform has been downloaded on your development machine, download the Terraform scripts here: `wget https://objectstorage.us-ashburn-1.oraclecloud.com/n/oracle-cloudnative/b/GrafanaTerraform/o/terraform_grafana.tar && tar -xvf terraform_grafana.tar`

`cd /terraform_grafana`. Initialize a new Terraform configuration with `terraform init`. This will result in the creation of two additional files in the directory: `terraform.tfstate` and `terraform.tfstate.backup`. 

Next run `terraform plan` to generate an execution plan. This is useful to validate your Terraform script prior to applying it to your environment. The command will output the end state of your environment. 

![Screen Shot 2018-12-17 at 3.59.38 PM](images/Screen%20Shot%202018-12-17%20at%203.59.38%20PM.png)

Run `terraform apply` and enter the same availability domain and dynamic group name variables used in the previous step. This will build the infrastructure specified in the execution plan. The following three actions will take place: 

1. The script will first create a [dynamic group](https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingdynamicgroups.htm) used to group virtual machine or bare metal compute instances as “principals” (similar to user groups). This will be named according to the variable you passed in during the previous step. 

   ![Screen Shot 2018-12-17 at 4.01.34 PM](images/Screen%20Shot%202018-12-17%20at%204.01.34%20PM.png)

2. Next, a [policy](https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm) named “grafana_policy” will be created in the root compartment of your tenancy to permit instances in the dynamic group to make API calls against Oracle Cloud Infrastructure services.

   ![Screen Shot 2018-12-17 at 4.01.47 PM](images/Screen%20Shot%202018-12-17%20at%204.01.47%20PM.png)

3. The script will then provision a [compute](https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm) instance with the name TFInstance0 in the compartment of your choice connected to the specified Virtual Cloud Network. After the instance is provisioned, the script will download and install Grafana and the Grafana OCI plugin. The plugin is stored in [Oracle Object Storage](https://docs.cloud.oracle.com/iaas/Content/Object/Concepts/objectstorageoverview.htm). 

![Screen Shot 2018-12-17 at 4.04.42 PM](images/Screen%20Shot%202018-12-17%20at%204.04.42%20PM.png)

Once the Terraform script is finished running you will receive an "Apply complete!" notification. 

## Configure Grafana

The next step is to configure the plugin. To find the IP address of the newly-created instance, navigate to Compute > Instances > TFInstance0. The Public IP address is listed under the Primary VNIC Information section. Using the SSH key referenced in your Terraform variables, connect to Grafana via port forward by running `ssh opc@[Instance Public IP] -L 3000:localhost:3000`. 

Navigate to the Grafana homepage at http://localhost:3000.

![Screen Shot 2018-12-17 at 3.23.46 PM](images/Screen%20Shot%202018-12-17%20at%203.23.46%20PM.png)

Log in with the default username `admin` and the password `admin`. You will be prompted to change your password. Click **Skip** or **Save** to continue. ![Screen Shot 2018-12-17 at 3.23.54 PM](images/Screen%20Shot%202018-12-17%20at%203.23.54%20PM.png)

On the Home Dashboard click the gear icon on the left side of the page.

![Screen Shot 2018-12-17 at 3.24.02 PM](images/Screen%20Shot%202018-12-17%20at%203.24.02%20PM.png)

Click **Add data source**.

![Screen Shot 2018-12-17 at 3.24.13 PM](images/Screen%20Shot%202018-12-17%20at%203.24.13%20PM.png)

 Choose **oracle-oci-datasource** as your data source type.

![Screen Shot 2018-12-17 at 3.24.24 PM](images/Screen%20Shot%202018-12-17%20at%203.24.17%20PM.png)

Fill in your **Tenancy OCID**, **Default Region**, and **Environment**. Your **Default region** is the same as your home region listed in the **Tenancy Details** page. For **Environment** choose **OCI Instance**. 

Click **Save & Test** to return to the home dashboard.

![Screen Shot 2018-12-17 at 3.25.33 PM](images/Screen_Shot_2019-02-08_at_10.19.56_AM.png)

## Clean Up 

To remove the environment created by the Terraform script, run `terraform destroy` and pass in the same availability domain and dynamic group variables passed in during the apply step.

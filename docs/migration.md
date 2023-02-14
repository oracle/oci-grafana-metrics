## Migration Instructions for Grafana OCI Metrics Data Source Settings (User Principals and Single Tenancy mode only)

DO NOT USE this Migration if your OCI Metrics Data Source is configured to use the `Instance Principals` authentication method. This guide is intended for users who are using the `User Principal` authentication method ! In case your Data Source is using `Instance Principals` authentication method there is no need to migrate. 

This document describes the steps to migrate data from `.oci/config` file to the Grafana data source settings for the following variables:

* `user`
* `fingerprint`
* `tenancy`
* `region`

Additionally, we will copy the ID key stored in a file (whose path is stored in the `key_file` variable) into the Grafana data source settings.

### Prerequisites

Before starting the migration process, please make sure you have the following:

* Access to the Grafana instance where the data source settings will be updated.
* The `.oci/config` file containing the variables to be migrated (`user`,`fingerprint`,`tenancy`,`region`) and the ID key (`key_file`).

### Steps

1. Log in to your Grafana instance and go to the data source settings page.
2. Locate the OCI Metrics data source where you want to update the settings and click on the "Edit" button. Configuration parameters will look like the following:
![Datasource Empty](images/datasource_conf_empty.png)
3. Choose `local`as `Environment` and `single` as `Tenancy Mode`
4. In the data source settings page, look for the fields corresponding to the variables that need to be updated (`user`,`fingerprint`,`tenancy`,`region`). Update the fields with the values from the `.oci/config` file.
5. Locate the field for the ID key and update it with the contents of the file stored at the path specified in the `key_file` variable.
6. Save the changes to the data source settings. Configuration parameters will look then like the following:
![Datasource Filled](images/datasource_conf_filled.png)


### Conclusion

By following these steps, you should have successfully migrated the data from the file to the Grafana data source settings for the specified variables and the ID key. If you have any issues or questions, please refer to the Grafana documentation or seek assistance from the Grafana community.

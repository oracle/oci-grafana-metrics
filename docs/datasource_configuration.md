
# Configure Grafana using datasource.yaml

You can manage data sources in Grafana by adding YAML configuration files in the provisioning/datasources directory. Each config file can contain a list of datasources to add or update during startup. If the data source already exists, Grafana reconfigures it to match the provisioned configuration file.

The configuration file can also list data sources to automatically delete, called deleteDatasources. Grafana deletes the data sources listed in deleteDatasources before adding or updating those in the datasources list.

For more details how datasource configuration works, you may refer to the official Grafana documentation [here](https://grafana.com/docs/grafana/latest/administration/provisioning/).

## Configure Grafana using datasource.yaml for Instance Principals

Following parameters must be set:

* *jsonData*
	+ **profile0**: A string that specifies the profile name. This field has a default value of 'DEFAULT', which is also the only allowed value.
	+ **environment**: A string that specifies the environment type, either 'local' or 'OCI Instance'.
* **secureJsonData**
	+ **version**: An integer that sets the version of the data source. This field is ignored when creating a new data source and is used to compare versions when updating.

Note: The DEFAULT value for profile0 is mandatory, as it is the only allowed value.

Here is the representation of the JSON data in a tabular format:

**jsonData**

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| jsonData | profile0 | Profile name, set to 'DEFAULT' for single instance and user principal. |
| jsonData | environment | Environment in which the application will run, set to 'OCI Instance'. |



## Configure Grafana using datasource.yaml for User Principals in Single tenancy mode

Following parameters must be set:

**jsonData**

* **profile0**: The profile name, it must be set to 'DEFAULT'.
* **environment**: The environment in which the application will run, use 'local' for User Principals in single tenancy mode.
* **tenancymode**: The tenancy mode, which can be either 'single' (for single-tenant) or 'multi' (for multi-tenant). For User Principals in single tenancy mode use 'single'.
* **region0**: The region code.

**secureJsonData**

This section contains sensitive data that should be encrypted before being stored in the JSON file. It includes:

* **user0**: user OCID.
* **tenancy0**: tenancy OCID.
* **fingerprint0**: Fingerprint value, hash of the API PEM key.
* **privkey0**: API PEM key. The key is formatted as a multi-line string.

Here is the representation of the JSON data in a tabular format:

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| jsonData | environment | The environment in which the application will run, either 'local' or an OCI Instance. |
| jsonData | tenancymode | The tenancy mode, set to  'multitenancy', indicating that the application supports multiple tenants. |
| jsonData | profile0 | Profile name for the first tenancy, always set to 'DEFAULT'. |
| jsonData | region0 | Region code for the first tenancy, set to 'eu-zurich-1'. |


| **Section** | **Element** | **Description** |
| --- | --- | --- |
| secureJsonData | user0 | User ID for the first tenancy. |
| secureJsonData | tenancy0 | Tenancy ID for the first tenancy. |
| secureJsonData | fingerprint0 | Fingerprint value for the first tenancy. |
| secureJsonData | privkey0 | Private key for the first tenancy. |



## Configure Grafana using datasource.yaml for User Principals in Multi tenancy mode

Following parameters must be set:

*  *jsonData*
	+ **environment*: The environment in which the application will run, which can be either 'local' (for local development) or an OCI Instance.
	+ **tenancymode*: The tenancy mode, which is set to  'multitenancy', indicating that the application supports multiple tenants.
	+ *profile0* and *region0*: These are the profile name and region code for the first tenancy. In this case, they are set to 'DEFAULT' and 'eu-zurich-1', respectively.
	+ *profile1* and *region1*: These are the profile name and region code for the second tenancy. In this case, they are set to 'SWEDEN' and 'eu-stockholm-1', respectively.
	+ *profile2*, *region2*, *profile3*, *region3*, *profile4*, and *region4*: These are the profile names and region codes for the third to fifth tenancies. In this case, they are not specified.

*  *secureJsonData*
This section contains sensitive data that should be encrypted before being stored in the JSON file. It includes:

	+ *user0*, *tenancy0*, *fingerprint0*, and *privkey0*: These are the user ID, tenancy ID, fingerprint value, and private key for the first tenancy.
	+ *user1*, *tenancy1*, *fingerprint1*, and *privkey1*: These are the user ID, tenancy ID, fingerprint value, and private key for the second tenancy.
	+ *user2*, *tenancy2*, *fingerprint2*, and *privkey2*: These are the user ID, tenancy ID, fingerprint value, and private key for the third tenancy.
	+ *user3*, *tenancy3*, *fingerprint3*, and *privkey3*: These are the user ID, tenancy ID, fingerprint value, and private key for the fourth tenancy.
	+ *user4*, *tenancy4*, *fingerprint4*, and *privkey4*: These are the user ID, tenancy ID, fingerprint value, and private key for the fifth tenancy.
	+ *user5*, *tenancy5*, *fingerprint5*, and *privkey5*: These are the user ID, tenancy ID, fingerprint value, and private key for the sixth tenancy.

Here is the representation of the JSON data in a tabular format:

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| jsonData | environment | The environment in which the application will run, either 'local' or an OCI Instance. |
| jsonData | tenancymode | The tenancy mode, set to  'multitenancy', indicating that the application supports multiple tenants. |
| jsonData | profile0 | Profile name for the first tenancy, always set to 'DEFAULT'. |
| jsonData | region0 | Region code for the first tenancy, set to 'eu-zurich-1'. |
| jsonData | profile1 | Profile name for the second tenancy, set to 'SWEDEN'. |
| jsonData | region1 | Region code for the second tenancy, set to 'eu-stockholm-1'. |
| ... | ... | ... |
| jsonData | profile5 | Profile name for the sixth tenancy (not specified). |
| jsonData | region5 | Region code for the sixth tenancy (not specified). |

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| secureJsonData | user0 | User ID for the first tenancy. |
| secureJsonData | tenancy0 | Tenancy ID for the first tenancy. |
| secureJsonData | fingerprint0 | Fingerprint value for the first tenancy. |
| secureJsonData | privkey0 | Private key for the first tenancy. |
| secureJsonData | user1 | User ID for the second tenancy. |
| secureJsonData | tenancy1 | Tenancy ID for the second tenancy. |
| secureJsonData | fingerprint1 | Fingerprint value for the second tenancy. |
| secureJsonData | privkey1 | Private key for the second tenancy. |
| ... | ... | ... |
| secureJsonData | user5 | User ID for the sixth tenancy. |
| secureJsonData | tenancy5 | Tenancy ID for the sixth tenancy. |
| secureJsonData | fingerprint5 | Fingerprint value for the sixth tenancy. |
| secureJsonData | privkey5 | Private key for the sixth tenancy. |
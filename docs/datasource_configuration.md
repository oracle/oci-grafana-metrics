
# Configure Grafana using datasource.yaml

You can manage data sources in Grafana by adding YAML configuration files in the provisioning/datasources directory. Each config file can contain a list of datasources to add or update during startup. If the data source already exists, Grafana reconfigures it to match the provisioned configuration file.

The configuration file can also list data sources to automatically delete, called deleteDatasources. Grafana deletes the data sources listed in deleteDatasources before adding or updating those in the datasources list.

For more details how datasource configuration works, you may refer to the official Grafana documentation [here](https://grafana.com/docs/grafana/latest/administration/provisioning/).

## Configure Grafana using datasource.yaml for Instance Principals

Following parameters must be set:

* *jsonData*
	+ **profile0**: A string that specifies the profile name. This field has a default value of 'DEFAULT', which is also the only allowed value.
	+ **environment**: A string that specifies the environment type, either 'local' or 'OCI Instance'. Use 'OCI Instance' for instance principals.

Note: The DEFAULT value for profile0 is mandatory, as it is the only allowed value.

Here is the representation of the required JSON data in a tabular format:

**jsonData**

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| jsonData | profile0 | Profile name, **must** be set to 'DEFAULT' for single instance and user principal. |
| jsonData | environment | Environment in which the application will run, set to 'OCI Instance'. |

### Configuration example for Instance Principal

```
# Configuration file version
apiVersion: 1

# List of data sources to delete from the database.
deleteDatasources:
  - name: InstancePrincipal
    orgId: 1

# List of data sources to insert/update depending on what's
# available in the database.
datasources:
  # <string, required> Sets the name you use to refer to
  # the data source in panels and queries.
  - name: InstancePrincipal
    # <string, required> Sets the data source type.
    type: oci-metrics-datasource
    # <string, required> Sets the access mode, either
    # proxy or direct (Server or Browser in the UI).
    # Some data sources are incompatible with any setting
    # but proxy (Server).
    access: proxy
    # <int> Sets the organization id. Defaults to orgId 1.
    orgId: 1
    # <string> Sets a custom UID to reference this
    # data source in other parts of the configuration.
    # If not specified, Grafana generates one.
    # uid: my_unique_uid
    # <string> Sets the database user, if necessary.
    user:
    # <string> Sets the database name, if necessary.
    database:
    # <bool> Enables basic authorization.multitenancy
    basicAuth:
    # <string> Sets the basic authorization username.
    basicAuthUser:
    # <bool> Enables credential headers.
    withCredentials:
    # <bool> Toggles whether the data source is pre-selected
    # for new panels. You can set only one default
    # data source per organization.
    isDefault:
    # <map> Fields to convert to JSON and store in jsonData.
    jsonData:
      # profile name (use DEFAULT for single instance and user principal)
      profile0: 'DEFAULT'
      # environment: use local or OCI Instance
      environment: 'OCI Instance'
    # <map> Fields to encrypt before storing in jsonData.
    secureJsonData:
    # <int> Sets the version. Used to compare versions when
    # updating. Ignored when creating a new data source.
    version: 1
    # <bool> Allows users to edit data sources from the
    # Grafana UI.
    editable: false
```

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

Here is the representation of the required JSON data in a tabular format:

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| jsonData | environment | The environment in which the application will run, either 'local' or an OCI Instance. |
| jsonData | tenancymode | The tenancy mode, set to  'singl', indicating that the datasource is configured in signel tenancy mode. |
| jsonData | profile0 | Profile name for the first tenancy, **must** be always set to 'DEFAULT'. |
| jsonData | region0 | Region code for the first tenancy. |


| **Section** | **Element** | **Description** |
| --- | --- | --- |
| secureJsonData | user0 | User ID for the first tenancy. |
| secureJsonData | tenancy0 | Tenancy ID for the first tenancy. |
| secureJsonData | fingerprint0 | Fingerprint value for the first tenancy. |
| secureJsonData | privkey0 | Private key for the first tenancy. |

### Configuration example for User Principal in Single Tenancy mode

```
# Configuration file version
apiVersion: 1

# List of data sources to delete from the database.
deleteDatasources:
  - name: Single
    orgId: 1

# List of data sources to insert/update depending on what's
# available in the database.
datasources:
  # <string, required> Sets the name you use to refer to
  # the data source in panels and queries.
  - name: Single
    # <string, required> Sets the data source type.
    type: oci-metrics-datasource
    # <string, required> Sets the access mode, either
    # proxy or direct (Server or Browser in the UI).
    # Some data sources are incompatible with any setting
    # but proxy (Server).
    access: proxy
    # <int> Sets the organization id. Defaults to orgId 1.
    orgId: 1
    # <string> Sets a custom UID to reference this
    # data source in other parts of the configuration.
    # If not specified, Grafana generates one.
    # uid: my_unique_uid
    # <string> Sets the database user, if necessary.
    user:
    # <string> Sets the database name, if necessary.
    database:
    # <bool> Enables basic authorization.multitenancy
    basicAuth:
    # <string> Sets the basic authorization username.
    basicAuthUser:
    # <bool> Enables credential headers.
    withCredentials:
    # <bool> Toggles whether the data source is pre-selected
    # for new panels. You can set only one default
    # data source per organization.
    isDefault:
    # <map> Fields to convert to JSON and store in jsonData.
    jsonData:
      # profile name (use DEFAULT for single instance and user principal, always DEFAULT for first tenancy in multi tenancy mode)
      profile0: 'DEFAULT'
      # environment: use local or OCI Instance
      environment: 'local'
      # use single or multitenancy
      tenancymode: 'single'
      # region code
      region0: 'eu-zurich-1'
    # <map> Fields to encrypt before storing in jsonData.
    secureJsonData:
      user0: 'ocid1.user.oc1..xxx'
      tenancy0: 'ocid1.tenancy.oc1..xxx'
      fingerprint0: 'XXX:XXX'
      privkey0: | 
        -----BEGIN PRIVATE KEY-----
        MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDiwW4Pkz20vFPr
        ...
        -----END PRIVATE KEY-----
      # <string> Sets the database password, if necessary.
      password:
      # <string> Sets the basic authorization password.
      basicAuthPassword:
    # <int> Sets the version. Used to compare versions when
    # updating. Ignored when creating a new data source.
    version: 1
    # <bool> Allows users to edit data sources from the
    # Grafana UI.
    editable: false

```


## Configure Grafana using datasource.yaml for User Principals in Multi tenancy mode

Following parameters must be set:

*  *jsonData*
	+ **environment*: The environment in which the application will run, which can be either 'local' (for local development) or an OCI Instance.
	+ **tenancymode*: The tenancy mode, which is set to  'multitenancy', indicating that the application supports multiple tenants.
	+ *profile0* and *region0*: These are the profile name and region code for the first tenancy. Profile name for first tenancy **must** be set to 'DEFAULT'
	+ *profile1* and *region1*: These are the profile name and region code for the second tenancy. 
	+ *profile2*, *region2*, *profile3*, *region3*, *profile4*, and *region4*: These are the profile names and region codes for the third to fifth tenancies.

*  *secureJsonData*
This section contains sensitive data that should be encrypted before being stored in the JSON file. It includes:

	+ *user0*, *tenancy0*, *fingerprint0*, and *privkey0*: These are the user ID, tenancy ID, fingerprint value, and private key for the first tenancy.
	+ *user1*, *tenancy1*, *fingerprint1*, and *privkey1*: These are the user ID, tenancy ID, fingerprint value, and private key for the second tenancy.
	+ *user2*, *tenancy2*, *fingerprint2*, and *privkey2*: These are the user ID, tenancy ID, fingerprint value, and private key for the third tenancy.
	+ *user3*, *tenancy3*, *fingerprint3*, and *privkey3*: These are the user ID, tenancy ID, fingerprint value, and private key for the fourth tenancy.
	+ *user4*, *tenancy4*, *fingerprint4*, and *privkey4*: These are the user ID, tenancy ID, fingerprint value, and private key for the fifth tenancy.
	+ *user5*, *tenancy5*, *fingerprint5*, and *privkey5*: These are the user ID, tenancy ID, fingerprint value, and private key for the sixth tenancy.

Here is the representation of the required JSON data in a tabular format:

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| jsonData | environment | The environment in which the application will run, either 'local' or an OCI Instance. use "local" for multitenancy configuration|
| jsonData | tenancymode | The tenancy mode, set to  'multitenancy', indicating that the application supports multiple tenants. |
| jsonData | profile0 | Profile name for the first tenancy, always set to 'DEFAULT'. |
| jsonData | region0 | Region code for the first tenancy |
| jsonData | profile1 | Profile name for the second tenancy. |
| jsonData | region1 | Region code for the second tenancy. |
| ... | ... | ... |
| jsonData | profile5 | Profile name for the sixth tenancy. |
| jsonData | region5 | Region code for the sixth tenancy. |

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

### Configuration example for User Principal in Multi Tenancy mode

```
# Configuration file version
apiVersion: 1

# List of data sources to delete from the database.
deleteDatasources:
  - name: Multi
    orgId: 1

# List of data sources to insert/update depending on what's
# available in the database.
datasources:
  # <string, required> Sets the name you use to refer to
  # the data source in panels and queries.
  - name: Multi
    # <string, required> Sets the data source type.
    type: oci-metrics-datasource
    # <string, required> Sets the access mode, either
    # proxy or direct (Server or Browser in the UI).
    # Some data sources are incompatible with any setting
    # but proxy (Server).
    access: proxy
    # <int> Sets the organization id. Defaults to orgId 1.
    orgId: 1
    # <string> Sets a custom UID to reference this
    # data source in other parts of the configuration.
    # If not specified, Grafana generates one.
    # uid: my_unique_uid
    # <string> Sets the database user, if necessary.
    user:
    # <string> Sets the database name, if necessary.
    database:
    # <bool> Enables basic authorization.multitenancy
    basicAuth:
    # <string> Sets the basic authorization username.
    basicAuthUser:
    # <bool> Enables credential headers.
    withCredentials:
    # <bool> Toggles whether the data source is pre-selected
    # for new panels. You can set only one default
    # data source per organization.
    isDefault:
    # <map> Fields to convert to JSON and store in jsonData.
    jsonData:
      # environment: use local or OCI Instance
      environment: 'local'
      # use single or multitenancy    
      tenancymode: 'multitenancy' 
      # profile name for first tenancy, set always to DEFAULT
      profile0: 'DEFAULT'
      # region code for first tenancy
      region0: 'eu-zurich-1'
      # profile name for second tenancy
      profile1: 'SWEDEN'
      # region code for second tenancy
      region1: 'eu-stockholm-1'      
    # <map> Fields to encrypt before storing in jsonData.
    secureJsonData:
      # user OCID
      user0: 'ocid1.user.oc1..xxx'
      # tenancy OCID
      tenancy0: 'ocid1.tenancy.oc1..xxx'
      # fingerprint of the api key
      fingerprint0: 'xxxx:xxxxxxxa'
      # api pem key
      privkey0: | 
        -----BEGIN PRIVATE KEY-----
        MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDiwW4Pkz20vFPr
        ...
        -----END PRIVATE KEY-----
      # user OCID for second tenancy
      user1: 'ocid1.user.oc1..xxx'
      # tenancy ocid for second tenancy
      tenancy1: 'ocid1.tenancy.oc1..xxx'
      # fingerprint of the api key for second tenancy      
      fingerprint1: 'xxx:xxxxxxx'
      # api pem key for second tenancy
      privkey1: | 
        -----BEGIN PRIVATE KEY-----
        MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDiwW4Pkz20vFPr
        ...
        -----END PRIVATE KEY-----
      # <string> Sets the database password, if necessary.
      password:
      # <string> Sets the basic authorization password.
      basicAuthPassword:
    # <int> Sets the version. Used to compare versions when
    # updating. Ignored when creating a new data source.
    version: 1
    # <bool> Allows users to edit data sources from the
    # Grafana UI.
    editable: false

```

# Configuring OCI Metrics Plugin Datasource using Grafana API

## Introduction

Configuring datasources in Grafana can be automated using the Grafana HTTP API. This approach allows for programmatic setup and management of datasources, which is particularly useful for large-scale deployments or when integrating Grafana into other systems.

To configure the OCI Metrics plugin datasource, you'll need to send a POST request to the Grafana API endpoint `/api/datasources`. The configuration details are sent in the request body as JSON. The specific parameters will vary depending on the authentication mode you're using: Single Tenancy, Multi-tenancy, or Instance Principal.

Before proceeding, ensure you have:

1. Grafana installed and running
2. Admin API key for authentication
3. OCI Metrics plugin installed in Grafana

Let's explore how to configure the OCI Metrics plugin datasource for each authentication mode.

## Configure Grafana using Grafana API for User Principals in Single tenancy mode

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

### Payload example for User Principals in Single tenancy mode

```
{
  "name": "api injected",
  "type": "oci-metrics-datasource",
  "editable": true,
  "access": "proxy",
  "jsonData": {
    "profile0": "DEFAULT",
    "environment": "local",
    "tenancymode": "single",
    "region0": "eu-zurich-1"
  },
  "secureJsonData": {
    "user0": "ocid1.user.oc1..XXXXXXX",
    "tenancy0": "ocid1.tenancy.oc1..XXXXXXXX",
    "fingerprint0": "62:d3:XXXXXXXXXXXX",
    "privkey0": "-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIXXXXXX=\n-----END PRIVATE KEY-----"
  }
}
```


## Configure Grafana using Grafana API for User Principals in Multi tenancy mode

Following parameters must be set:

*  *jsonData*
	+ **environment*: The environment in which the application will run, which can be either 'local' (for local development) or an OCI Instance.
	+ **tenancymode*: The tenancy mode, which is set to  'multitenancy', indicating that the application supports multiple tenants.
	+ *profile0* and *region0*: These are the profile name and region code for the first tenancy. Profile name for first tenancy **must** be set to 'DEFAULT'
	+ *profile1* and *region1*: These are the profile name and region code for the second tenancy. 
	+ *profile2*, *region2*, *profile3*, *region3*, *profile4*, and *region4*: These are the profile names and region codes for the third to fifth tenancies.

*  *secureJsonData*
This section contains sensitive data that should be encrypted before being stored in the JSON file. It includes:

	+ *user0*, *tenancy0*, *fingerprint0*, and *privkey0*: These are the user ID, tenancy ID, fingerprint value, and private key for the first tenancy.
	+ *user1*, *tenancy1*, *fingerprint1*, and *privkey1*: These are the user ID, tenancy ID, fingerprint value, and private key for the second tenancy.
	+ *user2*, *tenancy2*, *fingerprint2*, and *privkey2*: These are the user ID, tenancy ID, fingerprint value, and private key for the third tenancy.
	+ *user3*, *tenancy3*, *fingerprint3*, and *privkey3*: These are the user ID, tenancy ID, fingerprint value, and private key for the fourth tenancy.
	+ *user4*, *tenancy4*, *fingerprint4*, and *privkey4*: These are the user ID, tenancy ID, fingerprint value, and private key for the fifth tenancy.
	+ *user5*, *tenancy5*, *fingerprint5*, and *privkey5*: These are the user ID, tenancy ID, fingerprint value, and private key for the sixth tenancy.

### Payload example for User Principals in Multi tenancy mode

```
{
  "name": "api injected MULTI",
  "type": "oci-metrics-datasource",
  "editable": true,
  "access": "proxy",
  "jsonData": {
    "profile0": "DEFAULT",
    "environment": "local",
    "tenancymode": "multitenancy",
    "region0": "eu-zurich-1",
    "profile1": "SWEDEN",
    "region1": "eu-stockholm-1"    
  },
  "secureJsonData": {
    "user0": "ocid1.user.oc1..XXXXXX",
    "tenancy0": "ocid1.tenancy.oc1..XXXXXXXX",
    "fingerprint0": "62:d3:XXXXXXXXXXXX",
    "privkey0": "-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIXXXXXX=\n-----END PRIVATE KEY-----",
    "user1": "ocid1.user.oc1..aaaaaaaahsmfbxrirfsimdrdyouyuya26jvv5226hikv3gu3qbng63kwbmha",
    "tenancy1": "ocid1.tenancy.oc1..XXXXXXX",
    "fingerprint1": "62:d3:XXXXXXXXXXXX",
    "privkey1": "-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIXXXXXX=\n-----END PRIVATE KEY-----"    
  }
}
```

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| jsonData | environment | The environment in which the application will run, either 'local' or an OCI Instance. use "local" for multitenancy configuration|
| jsonData | tenancymode | The tenancy mode, set to  'multitenancy', indicating that the application supports multiple tenants. |
| jsonData | profile0 | Profile name for the first tenancy, always set to 'DEFAULT'. |
| jsonData | region0 | Region code for the first tenancy |
| jsonData | profile1 | Profile name for the second tenancy. |
| jsonData | region1 | Region code for the second tenancy. |
| ... | ... | ... |
| jsonData | profile5 | Profile name for the sixth tenancy. |
| jsonData | region5 | Region code for the sixth tenancy. |

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

Here is the representation of the required JSON data in a tabular format:

## Configure Grafana using Grafana API  for Instance Principals

Following parameters must be set:

* *jsonData*
	+ **profile0**: A string that specifies the profile name. This field has a default value of 'DEFAULT', which is also the only allowed value.
	+ **environment**: A string that specifies the environment type, either 'local' or 'OCI Instance'. Use 'OCI Instance' for instance principals.

Note: The DEFAULT value for profile0 is mandatory, as it is the only allowed value.

Here is the representation of the required JSON data in a tabular format:

**jsonData**

| **Section** | **Element** | **Description** |
| --- | --- | --- |
| jsonData | profile0 | Profile name, **must** be set to 'DEFAULT' for single instance and user principal. |
| jsonData | environment | Environment in which the application will run, set to 'OCI Instance'. |

### Payload example for Instance Principal

```
{
  "name": "api injected INSTANCE",
  "type": "oci-metrics-datasource",
  "editable": true,
  "access": "proxy",
  "jsonData": {
    "profile0": "DEFAULT",
    "environment": "OCI Instance"
  }
}
```

# Configure Grafana using datasource.yaml

You can manage data sources in Grafana by adding YAML configuration files in the provisioning/datasources directory. Each config file can contain a list of datasources to add or update during startup. If the data source already exists, Grafana reconfigures it to match the provisioned configuration file.

The configuration file can also list data sources to automatically delete, called deleteDatasources. Grafana deletes the data sources listed in deleteDatasources before adding or updating those in the datasources list.

For more details how datasource configuration works, you may refer to the official Grafana documentation [here](https://grafana.com/docs/grafana/latest/administration/provisioning/).

## Configure Grafana using datasource.yaml for Instance Principals

* *jsonData*
	+ **profile0**: A string that specifies the profile name. This field has a default value of 'DEFAULT', which is also the only allowed value.
	+ **environment**: A string that specifies the environment type, either 'local' or 'OCI Instance'.
* **secureJsonData**
	+ **version**: An integer that sets the version of the data source. This field is ignored when creating a new data source and is used to compare versions when updating.

Note: The DEFAULT value for profile0 is mandatory, as it is the only allowed value.
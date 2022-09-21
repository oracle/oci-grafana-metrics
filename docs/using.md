# Using Grafana with the Oracle Cloud Infrastructure Data Source 

Here are a few tools for you to explore after installing and configuring the Oracle Cloud Infrastructure Data Source for Grafana. Refer to [this page](https://docs.cloud.oracle.com/iaas/Content/Monitoring/Concepts/monitoringoverview.htm) for more information about how to use the Monitoring Service. 

## Query Editor

The query editor can be used to create graphs of your Oracle Cloud Infrastructure resources.

On the Grafana Home Dashboard and click **New Dashboard**.

![Screen Shot 2018-12-17 at 3.26.01 PM](images/Screen%20Shot%202018-12-17%20at%203.26.01%20PM.png)

Choose **Graph** from the list of available dashboard types.

![Screen Shot 2018-12-17 at 3.26.18 PM](images/Screen%20Shot%202018-12-17%20at%203.26.18%20PM.png)

Click **Panel Title** and then **Edit** to add metrics to the dashboard.![Screen Shot 2018-12-17 at 3.26.26 PM](images/Screen%20Shot%202018-12-17%20at%203.26.26%20PM.png)

Choose the appropriate **Region**, **Compartment**, **Namespace**, **Metric**, and **Dimension** from the list of available options.

![Metrics Query Editor](images/MetricsPlugin-QueryEditor-Screenshot.png)

Click the save icon to save your graph.

At this stage, if the **metrics** pull down menu is not properly populating with options, you may need to navigate back to the OCI console add an additional matching rule to your Dynamic Group stating: `matching_rule = “ANY {instance.compartment.id = ‘${var.compartment_ocid}’}”`. After doing so, restart the Grafana server as the **sudo** user run `systemctl restart grafana-server` and reload the Grafana console. 

## Templating 

Templating provides the ability to dynamically switch the contents of graphs as seen in the example below. 

![templating](images/templating.gif)

In order to configure templating, click on the gear icon in the upper right corner of the dashboard creation page from the previous step. This will take you to the **Settings** page. Click the **Variables** tab and then click the **Add variable** button. 

![Screen Shot 2019-01-11 at 3.10.49 PM](images/Screen%20Shot%202019-01-11%20at%203.10.49%20PM.png)

Add the **region** variable to this page. Give the variable the name `region`, choose **OCI** from the list of data sources, and for **Query** enter `regions()`. 

![Screen Shot 2019-01-11 at 3.00.28 PM](images/Screen%20Shot%202019-01-11%20at%203.00.28%20PM.png)

The page will load a preview of values  available for that variable. Scroll down and click **Add** to create a template variable for regions. 

![Screen Shot 2019-01-13 at 11.11.50 AM](images/Screen%20Shot%202019-01-13%20at%2011.11.50%20AM.png)

Repeat the process for the following OCI variables: 

| Name            | Query                                                                     |
| --------------- | ------------------------------------------------------------------------- |
| region          | `regions()`                                                               |
| compartment     | `compartments()`                                                          |
| namespace       | `namespaces($region,$compartment)`                                        |
| resourcegroup   | `resourcegroups($region, $compartment, $namespace)`                                        |
| metric          | `metrics($region,$compartment, $namespace, $resourcegroup)`                                |
| dimensionKey    | `dimensions($region, $compartment, $namespace, $metric, $resourcegroup)`                     |
| dimensionValue  | `dimensionOptions($region,$compartment,$namespace,$metric,$dimensionKey,$resourcegroup)` |

All of the metrics plugin template variables only support a singleton value with the exception of the dimension options template variable. For the dimension options template variable, the Multi-value radio button in the template variable configuration can be enabled and a user can select multiple dimension values to use within the query. The metric plugin runs the defined query once for each dimension value selected for the dimension options template variable.

The final list of variables should look like this: 

![Metrics dashboard variables screenshot](images/metrics-dashboard-variables-screenshot.png)

In order for these variables be available to be dynamically changed in your query, edit your existing query, and under **metrics** select the newly created variables for **region**, **compartment**, **namespace**, and **metric** as seen in the image below. 

![Screen Shot 2019-01-11 at 3.19.51 PM](images/Screen%20Shot%202019-01-11%20at%203.19.51%20PM.png)



For windows and resolution,

You can use a custom or constant variable. 
To create a custom, select the variable type as custom.
Label(appears as the display name of the variable selected)
  
Custom variables provide the convenience of selecting a value
out of many values. Try replicating the data in 
the following screenshots for window and resolution.

Custom window variable 
![Custom variable for window](images/customWIndowVariableCreation.png)


Custom reoslution variable
![Custom variable for resolution](images/customResolution%20variable%20creation.png)

Choose the save button to save your dashboard ! DO NOT FORGET 



Now, on the top of panel, you would see windowLabel and resolutionLabel
You can select the value for the variables from here  and use them in
the dropdowns window and resolution  below the panel

![Custom variable window dropdown](images/WithCustomDropDown.png)
 

### Dimensions

Dimensions can be used to add specificity to your graphs. To use dimensions create a new graph or navigate to an existing one and click the **Metrics** tab. After selecting your variables click the **+** next to **Dimensions** and select one of the tag filters from the list. For example, select `availabilityDomain` from the list. Next, click **select value** and choose from the newly populated list of dimension values. If you chose `availabilityDomain` as your tag filter, you should see dimension values corresponding to the availability domains in which you currently have services provisioned, for example `US-ASHBURN-AD-1`. 



![Screen Shot 2019-02-14 at 12.03.26 PM](images/Screen%20Shot%202019-02-14%20at%2012.03.26%20PM.png)


### Metric Label Customization

When the metric plugin performs a query defined in a data panel, the query results contain the queried metric name, the values of the metric in the specified time range, and the values of the dimensions associated with the resource to which the metric pertains. The plugin by default generates a label for each of the metrics returned in the query results of the form:

    <Metric Name>[<Dimension 1 value> | <Dimension 2 value> | ... <Dimension N value>]

where each of these placeholders is replaced by the relevant metric name and dimension values for a given metric. This label is what is shown in the panel legend to distinguish each metric. Metric dimensions that contain a resource OCID, either the ID of the resource with which the metric is associated or an OCI resource related to the resource with which the metric is associated, contain the full resource OCID.

Many OCI metrics have quite a few dimensions, as a result the default label in a data panel legend for such metrics may be quite long. The metrics plugin provides a means for you to customize the presentation of metrics labels in a data panel legend via the **Legend Format** field in the metrics query editor as shown in the following image.

![Explore Metrics Query Editor](images/MetricsPlugin-ExploreMetricsQueryEditor-Screenshot.png)

The Legend Format field for a metrics query can contain any literal text sequences (printable characters only) along with any number of the following placeholders. 

| Placeholder       | Value that will replace the placeholder           |
| ----------------- | ------------------------------------------------- |
| {{metric}}        | The name of metric	                            |
| {{dimensionName}} | The value of the specified metric dimension name  |

When the Legend Format field contains a defined format, the metrics plugin will generate a label for each metric that follows the defind format where each of the referenced placeholders is replaced by the relevant value for the metric. Any placeholders (or other text) in the legend format that do not line up with one of these placeholders will be unchanged. Note that placeholder labels are treated as case sensitive.

Examples of custom legend formats that could be defined for metrics associated with OCI resources include:

| Metric Namespace  | Example Metric   | Example Legend/Label Format                     | Example Resulting Metric Label    |
| ----------------- | ---------------- | ----------------------------------------------- | --------------------------------- |
| oci_computeagent  | CpuUtilization   | {{metric}} - {{resourceDisplayName}}            | CpuUtilization - InstanceName-123 |
| oci_computeagent  | DiskBytesWritten | {{availabilityDomain} - {{resourceDisplayName}} | bwAI:PHX-AD-2 - InstanceName-456  |
| oci_lbaas	        | BytesReceived    | {{lbName}} / {{backendSetName}}                 | myLB / backendset1                |
| oci_objectstorage	| ObjectCount      | {{tier}} ^ {{resourceDisplayName}}              | NORMAL ^ myBucketName             |
| oci_filestorage   | FileSystemUsage  | resourceId={{resourceId}}                       | resourceId=ocid1.filesystem....   |

## Custom Metrics and Namespaces

Oracle Cloud Infrastructure allows you to create [custom metrics namespaces](https://docs.cloud.oracle.com/iaas/Content/Monitoring/Tasks/publishingcustommetrics.htm), which can be used to ingest data from sources in addition to the native Oracle Cloud Infrastructure resources available by default. For example, an application could be instrumented to gather statistics about individual operations. The resource posting custom metrics must be able to authenticate to Oracle Cloud Infrastructure using either using the Oracle Cloud Infrastructure CLI authentication mentioned above or using [instance principals](https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/callingservicesfrominstances.htm). In the example below you can see the option to select `custom_namespace` from the **Namespace** drop down. 

![Screen_Shot_2019-02-15_at_2.53.37_PM](images/Screen_Shot_2019-02-15_at_2.53.37_PM.png)

You can also see two custom metrics `CustomMetric` and `CustomMetric2` from the **Metric** dropdown. 

![Screen_Shot_2019-02-15_at_2.59.47_PM](images/Screen_Shot_2019-02-15_at_2.59.47_PM.png)


----

Now, variables can be used in  window and resolution drop downs

#### MQL editor

Click the pencil icon to  add or edit the query in editor mode


For now, only the following fields can be updated in the query editor mode,
1. Metric
2. Window
3. Aggregation 

All other fields must be selected using the drop downs.
Use the pencil icon to toggle the editor mode.

Example:
Try pasting this ```AllRequests[1m].max()```

[1m] represents the value of window. Similarly, the value can be set to 2m, 3m , 1h etc
where the 1st part reprsents a number the second part represents the time denomination. 



| symbol            | Meaning                                                                     |
| --------------- | ------------------------------------------------------------------------- |
| s          |  second                                                           |
| m          |  minute                                                           |
| h          |  hour                                                           |

Note : Only lower cases are allowed for the above



###### Query with variables 

Variables can be used for window selection ```AllRequests[$window].max()```

Query with 'auto'.

Auto can be used to select the window automatically depending upon the time range ,
selected. ```AllRequests[auto].count()```


---- 

### Auto explained
Auto option is available for window and resolution. This helps the user to 
get an automatic configuration for window and resolution based on the following conditions. 

If the user selects a time range

1. Less than or equal to  7 days ->  window will be 1m and resolution will be 1 min
2. Less than or equal to  30 days and more than 7 days ->   window will be 5m and resolution will be 5 min.
3. More than 30 days -> a window will be 1h and resolution will be 1 h   
 

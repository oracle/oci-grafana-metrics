import React, { useState } from 'react';
import { InlineField, InlineFieldRow, FieldSet, SegmentAsync, AsyncMultiSelect, Input } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';
import { OCIDataSource } from './datasource';
import { OCIDataSourceOptions, AggregationOptions, IntervalOptions, OCIQuery, QueryPlaceholder } from './types';
import QueryModel from './query_model';
import { TenancyChoices } from './config.options';

type Props = QueryEditorProps<OCIDataSource, OCIQuery, OCIDataSourceOptions>;


export const QueryEditor: React.FC<Props> = (props) => {
  const { query, datasource, onChange, onRunQuery } = props;
  const tmode = datasource.getJsonData().tenancymode;
  const [hasTenancyDefault, setHasTenancyDefault] = useState(false);
  const [hasLegacyCompartment, setHasLegacyCompartment] = useState(false);
  const [tenancyValue, setTenancyValue] = useState(query.tenancyName);
  const [regionValue, setRegionValue] = useState(query.region);
  const [compartmentValue, setCompartmentValue] = useState(query.compartmentName);
  const [namespaceValue, setNamespaceValue] = useState(query.namespace);
  const [resourceGroupValue, setResourceGroupValue] = useState(query.resourceGroup);
  const [metricValue, setMetricValue] = useState(query.metric);
  // const [aggregationValue, setaggregationValue] = useState(query.aggregation);
  const [intervalValue, setIntervalValue] = useState(query.intervalLabel);
  const [legendFormatValue, setLegendFormatValue] = useState(query.legendFormat);  
  

  const onApplyQueryChange = (changedQuery: OCIQuery, runQuery = true) => {
    if (runQuery) {        
      const queryModel = new QueryModel(changedQuery, getTemplateSrv());
      // for metrics
      if (datasource.isVariable(String(query.metric))) {
        let { [String(query.metric)]: var_metric } = datasource.interpolateProps({ [String(query.metric)]: query.metric });
        console.log("OOO var_metric "+var_metric)
        if (var_metric !== "") { 
          query.metric = var_metric
        }
      } else {
        console.log("OOOelse var_metric "+query.metric)
      } 
      console.log("OOO changedQuery.queryText "+changedQuery.queryText)       
      if (queryModel.isQueryReady()) {
    
        changedQuery.queryText = queryModel.buildQuery(String(query.metric));

        onChange({ ...changedQuery });
        onRunQuery();
      }
    } else {
      onChange({ ...changedQuery });
    }
  };

  const init = () => {
    let initialDimensions: any = [];
    let initialTags: any = [];

    if (query.dimensionValues !== undefined && query.dimensionValues?.length > 0) {
      for (const eachDimension of query.dimensionValues) {
        let indexToSplit = eachDimension.lastIndexOf('=');
        let key = eachDimension.substring(0, indexToSplit);
        let val = eachDimension.substring(indexToSplit + 2, eachDimension.length - 1);

        initialDimensions.push({
          label: key + ' > ' + val,
          value: key + '="' + val + '"',
        });
      }
    }

    if (query.tagsValues !== undefined && query.tagsValues?.length > 0) {
      for (const eachTag of query.tagsValues) {
        let indexToSplit = eachTag.lastIndexOf('.');
        let key = eachTag.substring(0, indexToSplit);
        let val = eachTag.substring(indexToSplit + 1);

        initialTags.push({
          label: key + ' > ' + val,
          value: key + '=' + val,
        });
      }
    }

    return [initialDimensions, initialTags];
  };

  // const [initialDimensions, initialTags] = init();
  const [initialDimensions] = init();

  const [dimensionValue, setDimensionValue] = useState<Array<SelectableValue<string>>>(initialDimensions);
  // const [tagValue, setTagValue] = useState<Array<SelectableValue<string>>>(initialTags);
  // const [groupValue, setGroupValue] = useState<Array<SelectableValue<string>>>([]); 

  // Appends all available template variables to options used by dropdowns
  const addTemplateVariablesToOptions = (options: Array<SelectableValue<string>>) => {
    getTemplateSrv()
      .getVariables()
      .forEach((item) => {
        options.push({
          label: `$${item.name}`,
          value: `$${item.name}`,
        });
      });
    return options;
  }

  // fetch the tenancies from tenancies files, with name as key and ocid as value
  const getTenancyOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    options = addTemplateVariablesToOptions(options)
    const response = await datasource.getTenancies();
    if (response) {
      response.forEach((item: any) => {
        const sv: SelectableValue<string> = {
          label: item.name,
          value: item.ocid,
        };
        options.push(sv);
      });
    }
    return options;
  };

  const getCompartmentOptions = async () => {
    // const existingCompartmentsResponse = query.compartments;
    // if (query.namespace !== undefined) {
      let options: Array<SelectableValue<string>> = [];
      options = addTemplateVariablesToOptions(options)
      const response = await datasource.getCompartments(query.tenancyOCID);
      if (response) {
        response.forEach((item: any) => {
          const sv: SelectableValue<string> = {
            label: item.name,
            value: item.ocid,
          };
          options.push(sv);
        });
      }
      return options;
    // }
    // if (existingCompartmentsResponse) {
    //   return (existingCompartmentsResponse);
    // }
    // return [];
  };

  const getSubscribedRegionOptions = async () => {
    // const existingRegionsResponse = query.regions;
    // if (query.namespace !== undefined) {
      let options: Array<SelectableValue<string>> = [];
      options = addTemplateVariablesToOptions(options)
      const response = await datasource.getSubscribedRegions(query.tenancyOCID);
      if (response) {
        response.forEach((item: string) => {
          const sv: SelectableValue<string> = {
            label: item,
            value: item,
          };
          options.push(sv);
        });
      }
      return options;
    // }
    // if (existingRegionsResponse) {
    //   return (existingRegionsResponse);
    // }
    // return [];
  };

  const getNamespaceOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    options = addTemplateVariablesToOptions(options)
    const response = await datasource.getNamespacesWithMetricNames(
      query.tenancyOCID,
      query.compartmentOCID,
      query.region
    );
    if (response) {
      response.forEach((item: any) => {
        const sv: SelectableValue<string> = {
          label: item.namespace,
          value: item.metric_names,
        };
        options.push(sv);
      });
    }
    return options;
  };

  const getResourceGroupOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    options = addTemplateVariablesToOptions(options)
    const response = await datasource.getResourceGroupsWithMetricNames(
      query.tenancyOCID,
      query.compartmentOCID,
      query.region,
      query.namespace
    );
    if (response) {
      response.forEach((item: any) => {
        const sv: SelectableValue<string> = {
          label: item.resource_group,
          value: item.metric_names,
        };
        options.push(sv);
      });
    }
    return options;
  };

  // const getMetricOptions = async () => {
  //   let options: Array<SelectableValue<string>> = [];
  //   options = addTemplateVariablesToOptions(options)
  //   console.log("OOO0000 var_metric "+query.metricNames)
  //   const response = query.metricNames || [];
  //   console.log("OOO var_metric "+response)
  //   console.log("OOO111111 var_metric "+response)    
  //   response.forEach((item: any) => {
  //     const sv: SelectableValue<string> = {
  //       label: item,
  //       value: item,
  //     };
  //     console.log("OOO222 var_metric "+sv.value)
  //     options.push(sv);
  //   });
  //   return options;
  // };

  const getMetricOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    options = addTemplateVariablesToOptions(options)
  //   console.log("OOO0000 var_metric "+query.metricNames)
    // const response = query.metricNames || [];
    // console.log("OOO var_metric "+response)
    const response = await datasource.getResourceGroupsWithMetricNames(
      query.tenancyOCID,
      query.compartmentOCID,
      query.region,
      query.namespace
    );
    if (response) {
      response.forEach((item: any) => {
        item.metric_names.forEach((ii: any) => {
          const sv: SelectableValue<string> = {
            label: ii,
            value: ii,
          };
          console.log("OOO111111 var_metric "+ii)    
          options.push(sv);
        });
      });
    }
    return options;
  };

  const getAggregationOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    AggregationOptions.forEach((item: any) => {
      const sv: SelectableValue<any> = {
        label: item.label,
        value: item.value,
      };
      options.push(sv);
    });
    return options;
  };

  const getIntervalOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    IntervalOptions.forEach((item: any) => {
      const sv: SelectableValue<any> = {
        label: item.label,
        value: item.value,
      };
      options.push(sv);
    });
    return options;
  };

  const getDimensionOptions = () => {
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const response = await datasource.getDimensions(
          query.tenancyOCID,
          query.compartmentOCID,
          query.region,
          query.namespace,
          query.metric
        );
        const result = response.map((res: any) => {
          return {
            label: res.key,
            value: res.key,
            options: res.values.map((val: any) => {
              return { label: res.key + ' > ' + val, value: res.key + '="' + val + '"' };
            }),
          };
        });
        resolve(result);
      }, 0);
    });
  };
  // const getTagOptions = () => {
  //   return new Promise<Array<SelectableValue<string>>>((resolve) => {
  //     setTimeout(async () => {
  //       const response = await datasource.getTags(
  //         query.tenancyOCID,
  //         query.compartmentOCID,
  //         query.compartmentName,
  //         query.region,
  //         query.namespace
  //       );
  //       const result = response.map((res: any) => {
  //         return {
  //           label: res.key,
  //           value: res.key,
  //           options: res.values.map((val: any) => {
  //             return { label: res.key + ' > ' + val, value: res.key + '=' + val };
  //           }),
  //         };
  //       });
  //       resolve(result);
  //     }, 0);
  //   });
  // };
  // const getGroupByOptions = () => {
  //   return new Promise<Array<SelectableValue<string>>>((resolve) => {
  //     setTimeout(async () => {
  //       const response = await datasource.getDimensions(
  //         query.tenancyOCID,
  //         query.compartmentOCID,
  //         query.region,
  //         query.namespace,
  //         query.metric
  //       );
  //       const result = response.map((res: any) => {
  //         return {
  //           label: res.key,
  //           value: res.key,
  //         };
  //       });
  //       resolve(result);
  //     }, 0);
  //   });
  // };


  const getTenancyDefault = async () => {
    let tname: string;
    let tvalue: string;
    tname = 'DEFAULT/';
    tvalue = 'DEFAULT/';
    onApplyQueryChange(
      {
        ...query,
        tenancyName: tname,
        tenancyOCID: tvalue,
        compartments: new Promise<Array<SelectableValue<string>>>((resolve) => {
          setTimeout(async () => {
            const response = await datasource.getCompartments(tvalue);
            const result = response.map((res: any) => {
              return { label: res.name, value: res.ocid };
            });
            resolve(result);
          }, 0);
        }),
        regions: await getSubscribedRegionOptions(),
      },
      false
    );
  };

  const onTenancyChange = async (data: any) => {
    let tname: string;
    let tvalue: string;
    if (tmode !== TenancyChoices.multitenancy) {
      tname = 'DEFAULT/';
      tvalue = 'DEFAULT/';
    } else {
      setTenancyValue(data);
      tname = data.label
      tvalue = data.value
    }
    onApplyQueryChange(
      {
        ...query,
        tenancyName: tname,
        tenancyOCID: tvalue,
        compartments: new Promise<Array<SelectableValue<string>>>((resolve) => {
          setTimeout(async () => {
            const response = await datasource.getCompartments(tvalue);
            const result = response.map((res: any) => {
              return { label: res.name, value: res.ocid };
            });
            resolve(result);
          }, 0);
        }),
        compartmentName: undefined,
        compartmentOCID: undefined,
        regions: await getSubscribedRegionOptions(),
        region: undefined,
        namespace: undefined,
        metric: undefined,
      },
      false
    );
  };

  const onCompartmentChange = (data: any) => {
    setCompartmentValue(data);
    onApplyQueryChange(
      {
        ...query,
        compartmentName: data.label,
        compartmentOCID: data.value,
        namespace: undefined,
        metric: undefined,
      },
      false
    );
  };

  const onRegionChange = (data: SelectableValue) => {
    // eslint-disable-next-line no-debugger
    debugger;
    // insert the value into the options (custom value is enabled)
    if (query.regions && data.__isNew__) {
      query.regions = [...query.regions, { label: data.label, value: data.value }]
    }
    setRegionValue(data.value);   
    onApplyQueryChange({ ...query, region: data.value, namespace: undefined, metric: undefined }, false);
  };

  const onNamespaceChange = (data: any) => {
    // new Promise<Array<SelectableValue<string>>>(() => {
    //   setTimeout(async () => {
    //     await datasource.getTags(
    //       query.tenancyOCID,
    //       query.compartmentOCID,
    //       query.compartmentName,
    //       query.region,
    //       data.label
    //     );
    //   }, 0);
    // });
    setNamespaceValue(data);  
    onApplyQueryChange(
      {
        ...query,
        namespace: data.label,
        metricNames: data.value,
        metricNamesFromNS: data.value,
        resourceGroup: undefined,
        metric: undefined,
      },
      false
    );
  };

  const onResourceGroupChange = (data: any) => {
    //setRGValue(data);
    let mn: string[] = data.value;
    if (data.label === 'NoResourceGroup') {
      mn = query.metricNamesFromNS || [];
    }
    setResourceGroupValue(data);

    onApplyQueryChange({ ...query, resourceGroup: data.label, metricNames: mn, metric: undefined }, false);
  };

  const onMetricChange = (data: any) => {
    setMetricValue(data);
    // if (datasource.isVariable(String(data.value))) {
    //   let { [String(data.value)]: var_metric } = datasource.interpolateProps({ [String(data.value)]: data.value });
    //   console.log("OOO var_metric "+var_metric)
    //   if (var_metric !== "") { 
    //     data.value = var_metric
    //   }      
    // } else {
    //   console.log("OOO var_metric "+data.value)
    // }      
    onApplyQueryChange({ ...query, metric: data.value });
  };
  const onAggregationChange = (data: any) => {
    onApplyQueryChange({ ...query, statisticLabel: data.label, statistic: data.value });
  };
  const onIntervalChange = (data: any) => {
    setIntervalValue(data);
    onApplyQueryChange({ ...query, intervalLabel: data.label, interval: data.value });
  };
  const onLegendFormatChange = (data: any) => {
    setLegendFormatValue(data);
    console.log("onLegendFormatChange "+data)
    onApplyQueryChange({ ...query, legendFormat: data });
    };
  const onDimensionChange = (data: any) => {
    const existingDVs = query.dimensionValues || [];
    let newDimensionValues: string[] = [];

    const incomingDataLength = data.length;
    const existingDataLength = existingDVs.length;

    if (incomingDataLength < existingDataLength) {
      data.map((incomingD: any) => {
        newDimensionValues.push(incomingD.value);
      });
    } else {
      data.map((incomingD: any) => {
        let entriesAdded = 0;
        let isSameKey = false;
        const incomingDV = incomingD.value;

        for (const existingDV of existingDVs) {
          // adding existing value back
          if (incomingDV === existingDV) {
            newDimensionValues.push(existingDV);
            entriesAdded++;
            break;
          }

          // skipping add if key is same
          if (incomingDV.split('=')[0] === existingDV.split('=')[0]) {
            isSameKey = true;
            break;
          }
        }

        // for new and unique keys
        if (!isSameKey && entriesAdded === 0) {
          newDimensionValues.push(incomingDV);
        }
      });
    }

    if (incomingDataLength === newDimensionValues.length) {
      setDimensionValue(data);
      onApplyQueryChange({ ...query, dimensionValues: newDimensionValues });
    } else {
      query.dimensionValues = newDimensionValues;
    }
  };
  // const onTagChange = (data: any) => {
  //   let newTagsValues: string[] = [];

  //   data.map((incomingT: any) => {
  //     newTagsValues.push(incomingT.value);
  //   });

  //   setTagValue(data);
  //   onApplyQueryChange({ ...query, tagsValues: newTagsValues });
  // };
  // const onGroupByChange = (data: any) => {
  //   setGroupValue(data);
  //   let selectedGroup: string = QueryPlaceholder.GroupBy;

  //   if (data !== null) {
  //     selectedGroup = data.value;
  //   }

  //   onApplyQueryChange({ ...query, groupBy: selectedGroup });
  // };


  if (tmode !== TenancyChoices.multitenancy && !hasTenancyDefault) {
    getTenancyDefault();
    setHasTenancyDefault(true);
  }

  if (query.compartment !== "" && !hasLegacyCompartment) {
    console.log("Legacy compartment is present: " + query.compartment)
    query.compartmentOCID = query.compartment
    query.compartmentName = query.compartment
    setCompartmentValue(query.compartment);
    setHasLegacyCompartment(true);
    // onApplyQueryChange(
    //   {
    //     ...query,
    //     compartmentName: query.compartment,
    //     compartmentOCID: query.compartment,
    //     namespace: undefined,
    //     metric: undefined,
    //   },
    //   false
    // );    
  }

  return (
    <>
      <FieldSet>
        <InlineFieldRow>
          {tmode === TenancyChoices.multitenancy && (
            <>
              <InlineField label="TENANCY" labelWidth={20}>
                <SegmentAsync
                  className="width-14"
                  allowCustomValue={false}
                  required={true}
                  loadOptions={getTenancyOptions}
                  value={tenancyValue}
                  placeholder={QueryPlaceholder.Tenancy}
                  onChange={(data) => {
                    onTenancyChange(data);
                  }}
                />
              </InlineField>
            </>
          )}
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="REGION" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={true}
              required={true}
              loadOptions={getSubscribedRegionOptions}
              value={regionValue}
              placeholder={QueryPlaceholder.Region}
              onChange={(data) => {
                onRegionChange(data);
              }}
            />
          </InlineField>
          <InlineField label="COMPARTMENT" labelWidth={20}>
            <SegmentAsync
              className="width-28"
              allowCustomValue={true}
              required={false}
              loadOptions={getCompartmentOptions}
              value={compartmentValue}
              placeholder={QueryPlaceholder.Compartment}
              onChange={(data) => {
                onCompartmentChange(data);
              }}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="NAMESPACE" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={true}
              loadOptions={getNamespaceOptions}
              value={namespaceValue}
              placeholder={QueryPlaceholder.Namespace}
              onChange={(data) => {
                onNamespaceChange(data);
              }}
            />
          </InlineField>
          <InlineField label="RESOURCE GROUP" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={false}
              loadOptions={getResourceGroupOptions}
              value={resourceGroupValue}
              placeholder={QueryPlaceholder.ResourceGroup}
              onChange={(data) => {
                onResourceGroupChange(data);
              }}
            />
          </InlineField>
          <InlineField label="METRIC" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={true}
              loadOptions={getMetricOptions}
              value={metricValue}
              placeholder={QueryPlaceholder.Metric}
              onChange={(data) => {
                onMetricChange(data);
              }}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="AGGREGATION" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={true}
              loadOptions={getAggregationOptions}
              value={query.statisticLabel || AggregationOptions[0].label}
              placeholder={QueryPlaceholder.Aggregation}
              onChange={(data) => {
                onAggregationChange(data);
              }}
            />
          </InlineField>
          <InlineField label="INTERVAL" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={true}
              loadOptions={getIntervalOptions}
              // value={query.intervalLabel || IntervalOptions[0].label}
              value={intervalValue}
              placeholder={QueryPlaceholder.Interval}
              onChange={(data) => {
                onIntervalChange(data);
              }}
            />
          </InlineField>
          {/* <InlineField label="GROUP BY" labelWidth={20} tooltip="Start typing to see the options">
            <AsyncSelect
              className="width-14"
              isSearchable={true}
              defaultOptions={false}
              allowCustomValue={false}
              isClearable={true}
              backspaceRemovesValue={true}
              // loadOptions={getGroupByOptions}
              value={groupValue}
              placeholder={QueryPlaceholder.GroupBy}
              onChange={(data) => {
                onGroupByChange(data);
              }}
            />
          </InlineField> */}
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="DIMENSIONS" labelWidth={20} grow={true} tooltip="Start typing to see the options">
            <>
              <AsyncMultiSelect
                loadOptions={getDimensionOptions}
                isSearchable={true}
                defaultOptions={true}
                allowCustomValue={false}
                isClearable={true}
                closeMenuOnSelect={false}
                // placeholder={QueryPlaceholder.Dimensions}
                placeholder={""}
                value={dimensionValue}
                noOptionsMessage="Start typing to see values ..."
                onChange={(data) => {
                  onDimensionChange(data);
                }}
              />
            </>
          </InlineField>
        </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="LEGEND FORMAT" labelWidth={20} grow={true}>
              <>             
                <Input
                  className="width-30"
                  defaultValue={legendFormatValue}
                  onBlur={(event) => {
                    onLegendFormatChange(event.target.value);
                  }}                
                />
              </> 
          </InlineField>
        </InlineFieldRow>  
      </FieldSet>
    </>
  );
};


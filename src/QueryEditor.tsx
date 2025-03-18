/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import React, { useEffect, useState } from 'react';
import { InlineField, InlineFieldRow, FieldSet, SegmentAsync, AsyncMultiSelect, Input, TextArea, RadioButtonGroup } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';
import { OCIDataSource } from './datasource';
import { OCIDataSourceOptions, AggregationOptions, IntervalOptions, OCIQuery, QueryPlaceholder } from './types';
import QueryModel from './query_model';
import { TenancyChoices } from './config.options';

type Props = QueryEditorProps<OCIDataSource, OCIQuery, OCIDataSourceOptions>;

/**
 * QueryEditor Component
 *
 * This component provides a user interface for building and editing queries for the OCI data source.
 * It allows users to select various parameters such as tenancy, compartment, region, namespace, metric,
 * aggregation, interval, dimensions, and legend format. It also supports raw query input.
 */
export const QueryEditor: React.FC<Props> = (props) => {
  const { query, datasource, onChange, onRunQuery } = props;
  const tmode = datasource.getJsonData().tenancymode;
  const [hasLegacyCompartment, setHasLegacyCompartment] = useState(false);
  const [hasLegacyRawValue, setHasLegacyRawValue] = useState(false);
  const [hasLegacyTenancy, setHasLegacyTenancy] = useState(false);
  const [queryValue, setQueryValue] = useState(query.queryTextRaw);
  const [queryRawValue, setQueryRawValue] = useState(query.rawQuery);
  const [tenancyValue, setTenancyValue] = useState(query.tenancyName);
  const [regionValue, setRegionValue] = useState(query.region);
  const [compartmentValue, setCompartmentValue] = useState(query.compartmentName);
  const [namespaceValue, setNamespaceValue] = useState(query.namespace);
  const [resourcegroupValue, setResourceGroupValue] = useState(query.resourcegroup);
  const [metricValue, setMetricValue] = useState(query.metric);
  // const [aggregationValue, setaggregationValue] = useState(query.aggregation);
  const [intervalValue, setIntervalValue] = useState(query.intervalLabel);
  const [legendFormatValue, setLegendFormatValue] = useState(query.legendFormat);
  const [hasCalledGetTenancyDefault, setHasCalledGetTenancyDefault] = useState(false); 
  const editorModes = [
    { label: 'Raw Query', value: false },
    { label: 'Builder', value: true },
  ];  
  
  /**
   * onApplyQueryChange
   *
   * Applies changes to the query and optionally runs the query.
   *
   * @param changedQuery - The modified OCIQuery object.
   * @param runQuery - A boolean indicating whether to run the query after applying changes. Defaults to true.
   */  
  const onApplyQueryChange = (changedQuery: OCIQuery, runQuery = true) => {
    if (runQuery) {        
      const queryModel = new QueryModel(changedQuery, getTemplateSrv());

      onChange({ ...changedQuery });

      if (queryModel.isQueryReady()) {
        if (changedQuery.rawQuery === false){
          changedQuery.queryText = queryModel.buildQuery(String(changedQuery.queryTextRaw));
        } else {
          changedQuery.queryText = queryModel.buildQuery(String(changedQuery.metric));
        }
        onRunQuery();
      }

    } else {
      onChange({ ...changedQuery });
    }
  };

  /**
   * init
   *
   * Initializes the dimensions and tags based on the existing query.
   *
   * @returns An array containing the initial dimensions and tags.
   */
  const init = () => {
    let initialDimensions: any = [];
    let initialTags: any = [];

    if (query.dimensionValues !== undefined && query.dimensionValues?.length > 0) {
      for (const eachDimension of query.dimensionValues) {
        let indexToSplit = eachDimension.lastIndexOf('=');
        let key = eachDimension.substring(0, indexToSplit);
        let val = eachDimension.substring(indexToSplit + 2, eachDimension.length - 1);

        initialDimensions.push({
          label: key + ' - ' + val,
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
          label: key + ' - ' + val,
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

  /**
   * addTemplateVariablesToOptions
   *
   * Appends all available template variables to options used by dropdowns.
   *
   * @param options - The array of SelectableValue options to which template variables will be added.
   * @returns The updated array of SelectableValue options.
   */  
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

  
  /**
   * CustomInput Component
   *
   * A custom input field for Single Tenancy Mode, pre-filling the tenancy with "DEFAULT/".
   */  
  const CustomInput = ({ ...props }) => {
    const [isReady, setIsReady] = useState(false);
  
    useEffect(() => {
      if (!hasCalledGetTenancyDefault && isReady) {
        const getTenancyDefault = async () => {
          const tname = 'DEFAULT/';
          const tvalue = 'DEFAULT/';
          onApplyQueryChange({ ...query, tenancyName: tname, tenancy: tvalue }, false);
          setHasCalledGetTenancyDefault(true);
        };
        getTenancyDefault();
      }
    }, [isReady]);
  
    useEffect(() => {
      setIsReady(true);
    }, []);
  
    return <Input {...props} />;
  };

  /**
   * getTenancyOptions
   *
   * Fetches the available tenancies from the data source.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the tenancies.
   */  
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

  /**
   * getCompartmentOptions
   *
   * Fetches the available compartments for the selected tenancy from the data source.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the compartments.
   */  
  const getCompartmentOptions = async () => {
      let options: Array<SelectableValue<string>> = [];
      options = addTemplateVariablesToOptions(options)
      const response = await datasource.getCompartments(query.tenancy);
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

  /**
   * getSubscribedRegionOptions
   *
   * Fetches the subscribed regions for the selected tenancy from the data source.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the regions.
   */  
  const getSubscribedRegionOptions = async () => {
      let options: Array<SelectableValue<string>> = [];
      options = addTemplateVariablesToOptions(options)
      const response = await datasource.getSubscribedRegions(query.tenancy);
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
  };

  /**
   * getNamespaceOptions
   *
   * Fetches the available namespaces for the selected tenancy, compartment, and region from the data source.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the namespaces.
   */  
  const getNamespaceOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    options = addTemplateVariablesToOptions(options) 
    const response = await datasource.getNamespacesWithMetricNames(
      query.tenancy,
      query.compartment,
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

  /**
   * getResourceGroupOptions
   *
   * Fetches the available resource groups for the selected tenancy, compartment, region, and namespace from the data source.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the resource groups.
   */  
  const getResourceGroupOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    options = addTemplateVariablesToOptions(options)
    const response = await datasource.getResourceGroupsWithMetricNames(
      query.tenancy,
      query.compartment,
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

  /**
   * getMetricOptions
   *
   * Fetches the available metrics for the selected tenancy, compartment, region, and namespace from the data source.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the metrics.
   */  
  const getMetricOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    options = addTemplateVariablesToOptions(options)
    const response = await datasource.getResourceGroupsWithMetricNames(
      query.tenancy,
      query.compartment,
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
          options.push(sv);
        });
      });
    }
    return options;
  };

  /**
   * getAggregationOptions
   *
   * Returns the available aggregation options.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the aggregation options.
   */  
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

  /**
   * getIntervalOptions
   *
   * Returns the available interval options.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the interval options.
   */  
  const getIntervalOptions = async () => {
    let options: Array<SelectableValue<string>> = [];
    options = addTemplateVariablesToOptions(options);
    IntervalOptions.forEach((item: any) => {
      const sv: SelectableValue<any> = {
        label: item.label,
        value: item.value,
      };
      options.push(sv);
    });
    return options;
  };

  /**
   * getDimensionOptions
   *
   * Fetches the available dimensions for the selected tenancy, compartment, region, namespace, and metric from the data source.
   *
   * @returns A promise that resolves to an array of SelectableValue options representing the dimensions.
   */  
  const getDimensionOptions = () => {
    let templateOptions: Array<SelectableValue<string>> = [];
    templateOptions = addTemplateVariablesToOptions(templateOptions);
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const response = await datasource.getDimensions(
          query.tenancy,
          query.compartment,
          query.region,
          query.namespace,
          query.metric
        );
        const result = response.map((res: any) => {
          return {
            label: res.key,
            value: res.key,
            options: res.values.map((val: any) => {
              return { label: res.key + ' - ' + val, value: res.key + '="' + val + '"' };
            }),
          };
        });
        templateOptions.forEach((option) => {
          result.push({
            label: option.label,
            value: option.value,
            options: [{ label: option.label, value: option.value }]
          });
        });
        resolve(result);
      }, 0);
    });
  };
  
  
  // tags will be used in future release
  // const getTagOptions = () => {
  //   return new Promise<Array<SelectableValue<string>>>((resolve) => {
  //     setTimeout(async () => {
  //       const response = await datasource.getTags(
  //         query.tenancy,
  //         query.compartment,
  //         query.compartmentName,
  //         query.region,
  //         query.namespace
  //       );
  //       const result = response.map((res: any) => {
  //         return {
  //           label: res.key,
  //           value: res.key,
  //           options: res.values.map((val: any) => {
  //             return { label: res.key + ' - ' + val, value: res.key + '=' + val };
  //           }),
  //         };
  //       });
  //       resolve(result);
  //     }, 0);
  //   });
  // };


  /**
   * onTenancyChange
   *
   * Handles changes to the selected tenancy.
   *
   * @param data - The selected tenancy data.
   */
  const onTenancyChange = async (data: any) => {
    setTenancyValue(data);
    onApplyQueryChange(
      {
        ...query,
        tenancyName: data.label,
        tenancy: data.value,
        compartmentName: undefined,
        compartment: undefined,
        region: undefined,
        namespace: undefined,
        metric: undefined,
      },
      false
    );
  };

  /**
   * onRawQueryChange
   *
   * Handles changes to the raw query mode.
   *
   * @param data - The new raw query mode (true for builder, false for raw query).
   */  
  const onRawQueryChange = (data: boolean) => {
    setQueryRawValue(data);   
    onApplyQueryChange({ ...query, rawQuery: data }, false);
  };

  /**
   * onCompartmentChange
   *
   * Handles changes to the selected compartment.
   *
   * @param data - The selected compartment data.
   */  
  const onCompartmentChange = (data: any) => {
    setCompartmentValue(data);
    onApplyQueryChange(
      {
        ...query,
        compartmentName: data.label,
        compartment: data.value,
        namespace: undefined,
        metric: undefined,
      },
      false
    );
  };

  /**
   * onRegionChange 
   * 
   * Handles the change of the region selection.
   *
   * @param {SelectableValue} data - The selected region data.
   */  
  const onRegionChange = (data: SelectableValue) => {
    setRegionValue(data.value);   
    onApplyQueryChange({ ...query, region: data.value, namespace: undefined, metric: undefined }, false);
  };

  /**
   * onNamespaceChange
   * 
   * Handles the change of the namespace selection.
   *
   * @param {any} data - The selected namespace data.
   */  
  const onNamespaceChange = (data: any) => {
    // tags will be use in future release
    // new Promise<Array<SelectableValue<string>>>(() => {
    //   setTimeout(async () => {
    //     await datasource.getTags(
    //       query.tenancy,
    //       query.compartment,
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
        resourcegroup: undefined,
        metric: undefined,
      },
      false
    );
  };

  /**
   * onResourceGroupChange
   * 
   * Handles the change of the resource group selection.
   *
   * @param {any} data - The selected resource group data.
   */  
  const onResourceGroupChange = (data: any) => {
    let mn: string[] = data.value;
    setResourceGroupValue(data);
    onApplyQueryChange({ ...query, resourcegroup: data.label, metricNames: mn, metric: undefined }, false);
  };


  /**
   * onQueryTextChange
   * 
   * Handles the change of the query text.
   *
   * @param {any} data - The new query text.
   */  
  const onQueryTextChange = (data: any) => {
    setQueryValue(data);
    onApplyQueryChange({ ...query, queryTextRaw: data });
  };


  /**
   * onMetricChange
   * 
   * Handles the change of the metric selection.
   *
   * @param {any} data - The selected metric data.
   */
  const onMetricChange = (data: any) => {
    setMetricValue(data);
    onApplyQueryChange({ ...query, metric: data.value }, true);
  };

  /**
   * onAggregationChange
   * 
   * Handles the change of the aggregation selection.
   *
   * @param {any} data - The selected aggregation data.
   */  
  const onAggregationChange = (data: any) => {
    onApplyQueryChange({ ...query, statisticLabel: data.label, statistic: data.value });
  };


  /**
   * onIntervalChange
   * 
   * Handles the change of the interval selection.
   *
   * @param {any} data - The selected interval data.
   */  
  const onIntervalChange = (data: any) => {
    setIntervalValue(data);
    onApplyQueryChange({ ...query, intervalLabel: data.label, interval: data.value });
  };


  /**
   * onLegendFormatChange
   * 
   * Handles the change of the legend format.
   *
   * @param {any} data - The new legend format.
   */  
  const onLegendFormatChange = (data: any) => {
    setLegendFormatValue(data);
    onApplyQueryChange({ ...query, legendFormat: data });
  };


  /**
   * onDimensionChange
   * 
   * Handles the change of the dimension selection.
   *
   * @param {any} data - The selected dimension data.
   */  
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
  // Tags will be use in future release
  // const onTagChange = (data: any) => {
  //   let newTagsValues: string[] = [];

  //   data.map((incomingT: any) => {
  //     newTagsValues.push(incomingT.value);
  //   });

  //   setTagValue(data);
  //   onApplyQueryChange({ ...query, tagsValues: newTagsValues });
  // };

  // set tenancyName in case dashboard was created with version 4.x
  if (query.tenancy && !hasLegacyTenancy && !query.tenancyName) {
      query.tenancyName = query.tenancy;  
      setTenancyValue(query.tenancy);
      setHasLegacyTenancy(true);
  }

  // set compartmentName in case dashboard was created with version 4.x
  if (!query.compartmentName && query.compartment && !hasLegacyCompartment) {
    if (!query.tenancy && tmode === TenancyChoices.multitenancy) {
      return null;
    }
    datasource.getCompartments(query.tenancy).then(response => {
      if (response) {
        let found = false;
        response.forEach((item: any) => {
          if (!found && item.ocid === query.compartment) {
            found = true; 
            query.compartmentName = item.name;
          } else if (!found) {
            query.compartmentName = query.compartment;
          }           
        });
      } else {
          query.compartmentName = query.compartment;    
      }
      setCompartmentValue(query.compartmentName);
      setHasLegacyCompartment(true);
    });
}

  // set queryRawValue in case dashboard was created with version <= 5.0.0
  if (query.rawQuery === undefined && !hasLegacyRawValue) {
    setQueryRawValue(true);
    setHasLegacyRawValue(true);    
}

  return (        
    <>
      <FieldSet>
        <InlineFieldRow>
          {tmode === TenancyChoices.multitenancy && (
            <>
              <InlineField label="TENANCY" labelWidth={20}>
                <SegmentAsync
                  className="width-42"
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
          {tmode === TenancyChoices.single && (
            <>
        <InlineField label="TENANCY" labelWidth={20}>
          <CustomInput className="width-14" value={"DEFAULT/"} readOnly />
        </InlineField>
            </>
          )}
        <InlineField grow={true} className='container text-right'>
          <RadioButtonGroup
            options={editorModes}
            size="sm"
            value={queryRawValue}
            onChange={(data) => {
              onRawQueryChange(data);
            }}
          />
        </InlineField>             
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
              value={resourcegroupValue}
              placeholder={QueryPlaceholder.ResourceGroup}
              onChange={(data) => {
                onResourceGroupChange(data);
              }}
            />
          </InlineField>
          {query.rawQuery !== false && (
          <>          
          <InlineField label="METRIC" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={false}
              loadOptions={getMetricOptions}
              value={metricValue}
              placeholder={QueryPlaceholder.Metric}
              onChange={(data) => {
                onMetricChange(data);
              }}
            />
          </InlineField>
          </>
          )}          
        </InlineFieldRow>
        {query.rawQuery === false && (
            <>
            <InlineFieldRow>
              <InlineField
                      label="RAW QUERY"
                      labelWidth={20}
                      tooltip="type metric raw query"
                    >
                      <TextArea
                        type="text"
                        // className="width-70"
                        cols={80}
                        rows={4}
                        maxLength={16535}
                        defaultValue={queryValue}
                        onBlur={(event) => {
                          onQueryTextChange(event.target.value);
                        }}  
                        />
              </InlineField>
            </InlineFieldRow>              
            </>
          )}
        {query.rawQuery !== false && (
          <>                     
        <InlineFieldRow>
          <InlineField label="AGGREGATION" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={false}
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
              required={false}
              loadOptions={getIntervalOptions}
              // value={query.intervalLabel || IntervalOptions[0].label}
              value={intervalValue}
              placeholder={QueryPlaceholder.Interval}
              onChange={(data) => {
                onIntervalChange(data);
              }}
            />
          </InlineField>
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
          </>
            )}         
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


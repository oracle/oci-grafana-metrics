import React, { useState } from 'react';
import { InlineField, InlineFieldRow, FieldSet, SegmentAsync, AsyncMultiSelect, AsyncSelect } from '@grafana/ui';
import { QueryEditorProps, SelectableValue} from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';
import { OCIDataSource } from './datasource';
import { OCIDataSourceOptions, AggregationOptions, IntervalOptions, OCIQuery, QueryPlaceholder } from './types';
import QueryModel from './query_model';
import {TenancyChoices} from './config.options';

type Props = QueryEditorProps<OCIDataSource, OCIQuery, OCIDataSourceOptions>;


export const QueryEditor: React.FC<Props> = (props) => {
  const { query, datasource, onChange, onRunQuery } = props;
  const tmode = datasource.getJsonData().TenancyMode;


  const onApplyQueryChange = (changedQuery: OCIQuery, runQuery = true) => {
    if (runQuery) {
      const queryModel = new QueryModel(changedQuery, getTemplateSrv());
      if (queryModel.isQueryReady()) {
        changedQuery.queryText = queryModel.buildQuery();

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

  const [initialDimensions, initialTags] = init();
  const [dimensionValue, setDimensionValue] = useState<Array<SelectableValue<string>>>(initialDimensions);
  const [tagValue, setTagValue] = useState<Array<SelectableValue<string>>>(initialTags);
  const [groupValue, setGroupValue] = useState<Array<SelectableValue<string>>>([]);

  // fetch the tenancies from tenancies files, with name as key and ocid as value
  const getTenancyOptions = () => {
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const response = await datasource.getTenancies();
        const result = response.map((res: any) => {
          return { label: res.name, value: res.ocid };
        });
        resolve(result);
      }, 0);
    });
  };
  const getCompartmentOptions = () => {
    const existingCompartmentsResponse = query.compartments;

    if (query.namespace !== undefined) {
      return new Promise<Array<SelectableValue<string>>>((resolve) => {
        setTimeout(async () => {
          const response = await datasource.getCompartments(query.tenancyOCID);
          const result = response.map((res: any) => {
            return { label: res.name, value: res.ocid };
          });
          resolve(result);
        }, 0);
      });
    } else {
      return new Promise<Array<SelectableValue<string>>>((resolve) => {
        setTimeout(async () => {
          resolve(existingCompartmentsResponse);
        }, 0);
      });
    }
  };
  const getSubscribedRegionOptions = () => {
    const existingRegionsResponse = query.regions;

    if (query.namespace !== undefined) {
      return new Promise<Array<SelectableValue<string>>>((resolve) => {
        setTimeout(async () => {
          const response = await datasource.getSubscribedRegions(query.tenancyOCID);
          let result = response.map((res: any) => {
            return { label: res, value: res };
          });
          resolve(result);
        }, 0);
      });
    } else {
      return new Promise<Array<SelectableValue<string>>>((resolve) => {
        setTimeout(async () => {
          resolve(existingRegionsResponse);
        }, 0);
      });
    }
  };
  const getNamespaceOptions = () => {
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const response = await datasource.getNamespacesWithMetricNames(
          query.tenancyOCID,
          query.compartmentOCID,
          query.region
        );
        const result = response.map((res: any) => {
          return { label: res.namespace, value: res.metric_names };
        });
        resolve(result);
      }, 0);
    });
  };
  const getResourceGroupOptions = () => {
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const response = await datasource.getResourceGroupsWithMetricNames(
          query.tenancyOCID,
          query.compartmentOCID,
          query.region,
          query.namespace
        );
        const result = response.map((res: any) => {
          return { label: res.resource_group, value: res.metric_names };
        });
        resolve(result);
      }, 0);
    });
  };
  const getMetricOptions = () => {
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const response = query.merticNames || [];
        const result = response.map((res: any) => {
          return { label: res, value: res };
        });
        resolve(result);
      }, 0);
    });
  };
  const getAggregationOptions = () => {
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const result = AggregationOptions.map((res: any) => {
          return { label: res.label, value: res.value };
        });
        resolve(result);
      }, 0);
    });
  };
  const getIntervalOptions = () => {
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const result = IntervalOptions.map((res: any) => {
          return { label: res.label, value: res.value };
        });
        resolve(result);
      }, 0);
    });
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
  const getTagOptions = () => {
    return new Promise<Array<SelectableValue<string>>>((resolve) => {
      setTimeout(async () => {
        const response = await datasource.getTags(
          query.tenancyOCID,
          query.compartmentOCID,
          query.compartmentName,
          query.region,
          query.namespace
        );
        const result = response.map((res: any) => {
          return {
            label: res.key,
            value: res.key,
            options: res.values.map((val: any) => {
              return { label: res.key + ' > ' + val, value: res.key + '=' + val };
            }),
          };
        });
        resolve(result);
      }, 0);
    });
  };
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

  const onTenancyChange = (data: any) => {
    onApplyQueryChange(
      {
        ...query,
        tenancyName: data.label,
        tenancyOCID: data.value,
        compartments: new Promise<Array<SelectableValue<string>>>((resolve) => {
          setTimeout(async () => {
            const response = await datasource.getCompartments(data.value);
            const result = response.map((res: any) => {
              return { label: res.name, value: res.ocid };
            });
            resolve(result);
          }, 0);
        }),
        compartmentName: undefined,
        compartmentOCID: undefined,
        regions: new Promise<Array<SelectableValue<string>>>((resolve) => {
          setTimeout(async () => {
            const response = await datasource.getSubscribedRegions(data.value);
            let result = response.map((res: any) => {
              return { label: res, value: res };
            });
            resolve(result);
          }, 0);
        }),
        region: undefined,
        namespace: undefined,
        metric: undefined,
      },
      false
    );
  };
  const onCompartmentChange = (data: any) => {
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
  const onRegionChange = (data: any) => {
    onApplyQueryChange({ ...query, region: data.value, namespace: undefined, metric: undefined }, false);
  };
  const onNamespaceChange = (data: any) => {
    new Promise<Array<SelectableValue<string>>>(() => {
      setTimeout(async () => {
        await datasource.getTags(
          query.tenancyOCID,
          query.compartmentOCID,
          query.compartmentName,
          query.region,
          data.label
        );
      }, 0);
    });

    onApplyQueryChange(
      {
        ...query,
        namespace: data.label,
        merticNames: data.value,
        merticNamesFromNS: data.value,
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
      mn = query.merticNamesFromNS || [];
    }
    onApplyQueryChange({ ...query, resourceGroup: data.label, merticNames: mn, metric: undefined }, false);
  };
  const onMetricChange = (data: any) => {
    onApplyQueryChange({ ...query, metric: data.value });
  };
  const onAggregationChange = (data: any) => {
    onApplyQueryChange({ ...query, statisticLabel: data.label, statistic: data.value });
  };
  const onIntervalChange = (data: any) => {
    onApplyQueryChange({ ...query, intervalLabel: data.label, interval: data.value });
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
  const onTagChange = (data: any) => {
    let newTagsValues: string[] = [];

    data.map((incomingT: any) => {
      newTagsValues.push(incomingT.value);
    });

    setTagValue(data);
    onApplyQueryChange({ ...query, tagsValues: newTagsValues });
  };
  const onGroupByChange = (data: any) => {
    setGroupValue(data);
    let selectedGroup: string = QueryPlaceholder.GroupBy;

    if (data !== null) {
      selectedGroup = data.value;
    }

    onApplyQueryChange({ ...query, groupBy: selectedGroup });
  };

  return (
    <>
      <FieldSet>
        <InlineFieldRow>
        {tmode === TenancyChoices.multitenancy && (
          <>   
          <InlineField label="TENANCY" labelWidth={20} required={true}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={true}
              loadOptions={getTenancyOptions}
              value={query.tenancyName}
              placeholder={QueryPlaceholder.Tenancy}
              onChange={(data) => {
                onTenancyChange(data);
              }}
            />
          </InlineField>
          </>
        )}
          <InlineField label="REGION" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={true}
              loadOptions={getSubscribedRegionOptions}
              value={query.region}
              placeholder={QueryPlaceholder.Region}
              onChange={(data) => {
                onRegionChange(data);
              }}
            />
          </InlineField>          
          <InlineField label="COMPARTMENT" labelWidth={20}>
            <SegmentAsync
              className="width-14"
              allowCustomValue={false}
              required={false}
              loadOptions={getCompartmentOptions}
              value={query.compartmentName}
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
              value={query.namespace}
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
              value={query.resourceGroup}
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
              value={query.metric}
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
              value={query.intervalLabel || IntervalOptions[0].label}
              placeholder={QueryPlaceholder.Interval}
              onChange={(data) => {
                onIntervalChange(data);
              }}
            />
          </InlineField>
          <InlineField label="GROUP BY" labelWidth={20} tooltip="Start typing to see the options">
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
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="DIMENSIONS" labelWidth={20} grow={true} tooltip="Start typing to see the options">
            <>
              <AsyncMultiSelect
                loadOptions={getDimensionOptions}
                isSearchable={true}
                defaultOptions={false}
                allowCustomValue={false}
                isClearable={true}
                closeMenuOnSelect={false}
                placeholder={QueryPlaceholder.Dimensions}
                value={dimensionValue}
                onChange={(data) => {
                  onDimensionChange(data);
                }}
              />
            </>
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="TAGS" labelWidth={20} grow={true} tooltip="Start typing to see the options">
            <>
              <AsyncMultiSelect
                loadOptions={getTagOptions}
                isSearchable={true}
                defaultOptions={false}
                allowCustomValue={false}
                isClearable={true}
                closeMenuOnSelect={false}
                placeholder={QueryPlaceholder.Tags}
                value={tagValue}
                onChange={(data) => {
                  onTagChange(data);
                }}
              />
            </>
          </InlineField>
        </InlineFieldRow>
      </FieldSet>
    </>
  );
};

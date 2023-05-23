import React, { PureComponent } from 'react';
import { Input, Select, InlineField, FieldSet, InlineSwitch } from '@grafana/ui';
import {
  DataSourcePluginOptionsEditorProps,
  onUpdateDatasourceJsonDataOptionSelect,
  onUpdateDatasourceJsonDataOption,
  onUpdateDatasourceJsonDataOptionChecked,
} from '@grafana/data';
import { OCIDataSourceOptions, DefaultOCIOptions } from './types';
import {
  AuthProviders,
  regions,
  MultiTenancyChoices,
  TenancyChoices,
  AuthProviderOptions,
  MultiTenancyChoiceOptions,
  MultiTenancyModeOptions,
  TenancyChoiceOptions,
} from './config.options';

interface Props extends DataSourcePluginOptionsEditorProps<OCIDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  render() {
    const { options } = this.props;

    return (
      <FieldSet label="Connection Details">
        <InlineField
          label="Authentication Provider"
          labelWidth={28}
          tooltip="Specify which OCI credentials chain to use"
        >
          <Select
            className="width-30"
            value={options.jsonData.authProvider || ''}
            options={AuthProviderOptions}
            defaultValue={options.jsonData.authProvider}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'authProvider')(option);
            }}
          />
        </InlineField>

        <InlineField              
              label="Tenancy Mode"
              labelWidth={28}
              tooltip="Choose if want to enable multi-tenancy mode to fetch metrics accross multiple OCI tenancies"
            >
              <Select
                className="width-30"
                value={options.jsonData.TenancyChoice || ''}
                options={TenancyChoiceOptions}
                defaultValue={TenancyChoiceOptions[1]}
                onChange={(option) => {
                  onUpdateDatasourceJsonDataOptionSelect(this.props, 'TenancyChoice')(option);
                }}
              />
            </InlineField>
            <br></br>

{/* Instance Principals  */}
        {options.jsonData.authProvider === AuthProviders.OCI_INSTANCE && (
          <>
      <FieldSet label="Instance Principals Connection Details">
        <InlineField
          label="Default Region"
          labelWidth={28}
          tooltip="Specify the default Region for the tenancy"
        >
          <Select
            className="width-30"
            options={regions.map((region) => ({
              label: region,
              value: region,
              }))}
            defaultValue={options.jsonData.authProvider}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'defaultRegion')(option);
            }}
          />
        </InlineField>        

        <InlineField
          label="Default Region"
          labelWidth={28}
          tooltip="Specify the default Region for the tenancy"
        >
          <Input
            className="width-30"
            // css=""
            value={options.jsonData.defaultRegion || ''}
            required={true}
            onChange={onUpdateDatasourceJsonDataOption(this.props, 'defaultRegion')}
          />
        </InlineField>
        </FieldSet>
        </>
        )}


{/* User Principals  */}
  {options.jsonData.authProvider === AuthProviders.OCI_USER && (
              <>
      <FieldSet label="DEFAULT Connection Details">
      <InlineField
              label="Config Profile Name"
              labelWidth={28}
              tooltip="Config profile name. Default value is DEFAULT."
            >
              <Input
                className="width-30"
                // css=""
                placeholder={DefaultOCIOptions.ConfigProfile}
                value={options.jsonData.configProfile || DefaultOCIOptions.ConfigProfile}
                onChange={onUpdateDatasourceJsonDataOption(this.props, 'configProfile')}
              />
            </InlineField>
      <InlineField
          label="Region"
          labelWidth={28}
          tooltip="Specify the Region"
        >
          <Select
            className="width-30"
            options={regions.map((region) => ({
              label: region,
              value: region,
              }))}
            defaultValue={options.jsonData.authProvider}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'defaultRegion')(option);
            }}
          />
        </InlineField>
        

      </FieldSet>
      </>
        )}  


        {options.jsonData.TenancyChoice === TenancyChoices.multitenancy && (
          <>                          
        <InlineField
          label="Base Tenancy Name"
          labelWidth={28}
          tooltip="Specify the tenancy name where user profile is associated or instance is deployed"
        >
          <Input
            className="width-30"
            // css=""
            value={options.jsonData.tenancyName || ''}
            required={true}
            onChange={onUpdateDatasourceJsonDataOption(this.props, 'tenancyName')}
          />
        </InlineField>
        </>
        )}         
        <InlineField
          label="Authentication Provider"
          labelWidth={28}
          tooltip="Specify which OCI credentials chain to use"
        >
          <Select
            className="width-30"
            value={options.jsonData.authProvider || ''}
            options={AuthProviderOptions}
            defaultValue={options.jsonData.authProvider}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'authProvider')(option);
            }}
          />
        </InlineField>
        {options.jsonData.authProvider === AuthProviders.OCI_USER && (
          <>
            <InlineField label="Config Path" labelWidth={28} tooltip="Config file path. Default path is ~/.oci/config.">
              <Input
                className="width-30"
                // css=""
                placeholder={DefaultOCIOptions.ConfigPath}
                value={options.jsonData.configPath || DefaultOCIOptions.ConfigPath}
                onChange={onUpdateDatasourceJsonDataOption(this.props, 'configPath')}
              />
            </InlineField>

            <InlineField
              label="Enable Multi-Tenancy"
              labelWidth={28}
              tooltip="Choose if want to enable multi-tenancy mode to fetch metrics accross multiple OCI tenancies"
            >
              <Select
                className="width-30"
                value={options.jsonData.multiTenancyChoice || ''}
                options={MultiTenancyChoiceOptions}
                defaultValue={MultiTenancyChoiceOptions[1]}
                onChange={(option) => {
                  onUpdateDatasourceJsonDataOptionSelect(this.props, 'multiTenancyChoice')(option);
                }}
              />
            </InlineField>
            {options.jsonData.multiTenancyChoice === MultiTenancyChoices.YES && (
              <>
                <InlineField
                  label="Mode of Multi-Tenancy"
                  labelWidth={28}
                  tooltip="The mode via which multi-tenancy will be used."
                >
                  <Select
                    className="width-30"
                    value={options.jsonData.multiTenancyMode || ''}
                    options={MultiTenancyModeOptions}
                    defaultValue={MultiTenancyModeOptions[0]}
                    onChange={(option) => {
                      onUpdateDatasourceJsonDataOptionSelect(this.props, 'multiTenancyMode')(option);
                    }}
                  />
                </InlineField>
                <InlineField
                  label="Tenancy List File"
                  labelWidth={28}
                  tooltip="File which will contain list of target tenancy information in the format '<tenancy_name>,<tenancy_ocid>'. Default path is ~/.oci/tenancies"
                >
                  <Input
                    className="width-30"
                    // css=""
                    placeholder={DefaultOCIOptions.MultiTenanciesFile}
                    value={options.jsonData.multiTenancyFile || DefaultOCIOptions.MultiTenanciesFile}
                    onChange={onUpdateDatasourceJsonDataOption(this.props, 'multiTenancyFile')}
                  />
                </InlineField>
              </>
            )}
          </>
        )}
        <InlineField
          label="Enable Customer Mapping"
          labelWidth={28}
          tooltip="Provide more metadata about resources under tenancy. Possible via two way. 1. Using oracle CMDB, 2. Using user provide customer mapping file"
        >
          <InlineSwitch
            className="width-30"
            // css=""
            value={!!options.jsonData.enableCMDB}
            defaultChecked={false}
            onChange={onUpdateDatasourceJsonDataOptionChecked(this.props, 'enableCMDB')}
          />
        </InlineField>
        {options.jsonData.enableCMDB === true && (
          <>
            <InlineField
              label="Enable Mapping via file (excel)"
              labelWidth={28}
              tooltip="Customer mapping excel for tenancy resource. It must be an excel file."
            >
              <InlineSwitch
                className="width-30"
                // css=""
                value={!!options.jsonData.enableCMDBUploadFile}
                defaultChecked={false}
                onChange={onUpdateDatasourceJsonDataOptionChecked(this.props, 'enableCMDBUploadFile')}
              />
            </InlineField>
          </>
        )}
      </FieldSet>
    );
  }
}

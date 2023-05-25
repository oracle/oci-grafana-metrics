import React, { PureComponent } from 'react';
import { Input, Select, InlineField, FieldSet, InlineSwitch, TextArea } from '@grafana/ui';
import {
  DataSourcePluginOptionsEditorProps,
  onUpdateDatasourceJsonDataOptionSelect,
  onUpdateDatasourceJsonDataOption,
  onUpdateDatasourceJsonDataOptionChecked,
  onUpdateDatasourceSecureJsonDataOption,
} from '@grafana/data';
import { OCIDataSourceOptions, DefaultOCIOptions } from './types';
import {
  AuthProviders,
  regions,
  // MultiTenancyChoices,
  TenancyChoices,
  AuthProviderOptions,
  // MultiTenancyChoiceOptions,
  // MultiTenancyModeOptions,
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
            value={options.jsonData.Environment || ''}
            options={AuthProviderOptions}
            defaultValue={options.jsonData.Environment}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'Environment')(option);
            }}
          />
        </InlineField>

        {options.jsonData.Environment === AuthProviders.OCI_USER && (
              <>
        <InlineField              
              label="Tenancy Mode"
              labelWidth={28}
              tooltip="Choose if want to enable multi-tenancy mode to fetch metrics accross multiple OCI tenancies"
            >
              <Select
                className="width-30"
                value={options.jsonData.TenancyMode || ''}
                options={TenancyChoiceOptions}
                defaultValue={options.jsonData.TenancyMode}
                onChange={(option) => {
                  onUpdateDatasourceJsonDataOptionSelect(this.props, 'TenancyMode')(option);
                }}
              />
            </InlineField>
            </>
        )}
            <br></br>

{/* Instance Principals  */}
        {options.jsonData.Environment === AuthProviders.OCI_INSTANCE && (
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
            defaultValue={options.jsonData.Environment}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'defaultRegion')(option);
            }}
          />
        </InlineField>
        </FieldSet>
        </>
        )}


{/* User Principals - Default tenancy */}
  {options.jsonData.Environment === AuthProviders.OCI_USER && (
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
                onChange={onUpdateDatasourceJsonDataOption(this.props, 'profile0')}
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
            defaultValue={options.jsonData.region0}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'region0')(option);
            }}
          />
        </InlineField>
        <InlineField
              label="User OCID"
              labelWidth={28}
              tooltip="User OCID"
            >
              <Input
                className="width-30"
                placeholder={options.secureJsonFields.user0 ? 'configured' : ''}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'user0')}
                />
      </InlineField>
      <InlineField
              label="Tenancy OCID"
              labelWidth={28}
              tooltip="Tenancy OCID"
            >
              <Input
                className="width-30"
                placeholder={options.secureJsonFields.tenancy0 ? 'configured' : ''}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'tenancy0')}
                />
      </InlineField>
      <InlineField
              label="Fingerprint"
              labelWidth={28}
              tooltip="Fingerprint"
            >
              <Input
                placeholder={options.secureJsonFields.fingerprint0 ? 'configured' : ''}
                className="width-30"
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'fingerprint0')}
                />
      </InlineField>
      <InlineField
              label="Private Key"
              labelWidth={28}
              tooltip="Private Key"
            >
              <TextArea
                type="text"
                className="width-30"
                placeholder={options.secureJsonFields.privkey0 ? 'configured' : ''}
                cols={20}
                rows={4}
                maxLength={4096}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'privkey0')}
                />
      </InlineField>

      </FieldSet>
      </>
      )}  


{/* User Principals - Multitenancy Tenancy 1*/}
        {options.jsonData.TenancyMode === TenancyChoices.multitenancy && (
          <>                          
      <FieldSet label="Tenancy-1 Connection Details">
      <InlineField
              label="Config Profile Name"
              labelWidth={28}
              tooltip="Config profile name. Default value is DEFAULT."
            >
              <Input
                className="width-30"
                onChange={onUpdateDatasourceJsonDataOption(this.props, 'profile1')}
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
            defaultValue={options.jsonData.region1}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'region1')(option);
            }}
          />
        </InlineField>
        <InlineField
              label="User OCID"
              labelWidth={28}
              tooltip="User OCID"
            >
              <Input
                className="width-30"
                placeholder={options.secureJsonFields.user1 ? 'configured' : ''}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'user1')}
                />
      </InlineField>
      <InlineField
              label="Tenancy OCID"
              labelWidth={28}
              tooltip="Tenancy OCID"
            >
              <Input
                className="width-30"
                placeholder={options.secureJsonFields.tenancy1 ? 'configured' : ''}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'tenancy1')}
                />
      </InlineField>
      <InlineField
              label="Fingerprint"
              labelWidth={28}
              tooltip="Fingerprint"
            >
              <Input
                className="width-30"
                placeholder={options.secureJsonFields.fingerprint1 ? 'configured' : ''}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'fingerprint1')}
                />
      </InlineField>
      <InlineField
              label="Private Key"
              labelWidth={28}
              tooltip="Private Key"
            >
              <TextArea
                type="text"
                className="width-30"
                placeholder={options.secureJsonFields.privkey1 ? 'configured' : ''}
                cols={20}
                rows={4}
                maxLength={4096}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'privkey1')}
                />
      </InlineField>
      <InlineField
          label="Add another Tenancy ?"
          labelWidth={28}
          tooltip="Add Another tenancy YES/NO"
        >
          <InlineSwitch
            className="width-30"
            defaultChecked={false}
            onChange={onUpdateDatasourceJsonDataOptionChecked(this.props, 'addon1')}
          />
        </InlineField>
      </FieldSet>
        </>
        )}

{/* User Principals - Multitenancy Tenancy 2*/}
        {options.jsonData.addon1 === true && (
          <>
      <FieldSet label="Tenancy-2 Connection Details">
      <InlineField
              label="Config Profile Name"
              labelWidth={28}
              tooltip="Config profile name."
            >
              <Input
                className="width-30"
                onChange={onUpdateDatasourceJsonDataOption(this.props, 'profile2')}
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
            defaultValue={options.jsonData.region2}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'region2')(option);
            }}
          />
        </InlineField>
        <InlineField
              label="User OCID"
              labelWidth={28}
              tooltip="User OCID"
              placeholder={options.secureJsonFields.user2 ? 'configured' : ''}
            >
              <Input
                className="width-30"
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'user2')}
                />
      </InlineField>
      <InlineField
              label="Tenancy OCID"
              labelWidth={28}
              tooltip="Tenancy OCID"
            >
              <Input
                className="width-30"
                placeholder={options.secureJsonFields.tenancy2 ? 'configured' : ''}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'tenancy2')}
                />
      </InlineField>
      <InlineField
              label="Fingerprint"
              labelWidth={28}
              tooltip="Fingerprint"
            >
              <Input
                className="width-30"
                placeholder={options.secureJsonFields.fingerprint2 ? 'configured' : ''}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'fingerprint2')}
                />
      </InlineField>
      <InlineField
              label="Private Key"
              labelWidth={28}
              tooltip="Private Key"
            >
              <TextArea
                type="text"
                className="width-30"
                placeholder={options.secureJsonFields.privkey2 ? 'configured' : ''}
                cols={20}
                rows={4}
                maxLength={4096}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'privkey2')}
                />
      </InlineField>
      <InlineField
          label="Add another Tenancy ?"
          labelWidth={28}
          tooltip="Add Another tenancy YES/NO"
        >
          <InlineSwitch
            className="width-30"
            defaultChecked={false}
            onChange={onUpdateDatasourceJsonDataOptionChecked(this.props, 'addon2')}
          />
        </InlineField>
      </FieldSet>
          </>
        )}

{/* User Principals - Multitenancy Tenancy 3*/}
{options.jsonData.addon1 === true && (
          <>
      <FieldSet label="Tenancy-3 Connection Details">
      <InlineField
              label="Config Profile Name"
              labelWidth={28}
              tooltip="Config profile name."
            >
              <Input
                className="width-30"
                onChange={onUpdateDatasourceJsonDataOption(this.props, 'profile3')}
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
            defaultValue={options.jsonData.region3}
            onChange={(option) => {
              onUpdateDatasourceJsonDataOptionSelect(this.props, 'region3')(option);
            }}
          />
        </InlineField>
        <InlineField
              label="User OCID"
              labelWidth={28}
              tooltip="User OCID"
              placeholder={options.secureJsonFields.user2 ? 'configured' : ''}
            >
              <Input
                className="width-30"
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'user3')}
                />
      </InlineField>
      <InlineField
              label="Tenancy OCID"
              labelWidth={28}
              tooltip="Tenancy OCID"
              placeholder={options.secureJsonFields.tenancy2 ? 'configured' : ''}
            >
              <Input
                className="width-30"
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'tenancy2')}
                />
      </InlineField>
      <InlineField
              label="Fingerprint"
              labelWidth={28}
              tooltip="Fingerprint"
              placeholder={options.secureJsonFields.fingerprint2 ? 'configured' : ''}
            >
              <Input
                className="width-30"
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'fingerprint3')}
                />
      </InlineField>
      <InlineField
              label="Private Key"
              labelWidth={28}
              tooltip="Private Key"
              placeholder={options.secureJsonFields.privkey2 ? 'configured' : ''}
            >
              <TextArea
                type="text"
                className="width-30"
                cols={20}
                rows={4}
                maxLength={4096}
                onChange={onUpdateDatasourceSecureJsonDataOption(this.props, 'privkey3')}
                />
      </InlineField>
      <InlineField
          label="Add another Tenancy ?"
          labelWidth={28}
          tooltip="Add Another tenancy YES/NO"
        >
          <InlineSwitch
            className="width-30"
            defaultChecked={false}
            onChange={onUpdateDatasourceJsonDataOptionChecked(this.props, 'addon3')}
          />
        </InlineField>
      </FieldSet>
          </>
        )}


      </FieldSet>
    );
  }
}

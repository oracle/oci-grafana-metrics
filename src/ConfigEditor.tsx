import React, { PureComponent } from 'react';
import { Input, Select, InlineField, FieldSet, InlineSwitch, FileDropzone } from '@grafana/ui';
import {
  DataSourcePluginOptionsEditorProps,
  onUpdateDatasourceJsonDataOptionSelect,
  onUpdateDatasourceJsonDataOption,
  onUpdateDatasourceJsonDataOptionChecked,
} from '@grafana/data';
import { OCIDataSourceOptions, DefaultOCIOptions } from './types';
import {
  AuthProviders,
  MultiTenancyChoices,
  AuthProviderOptions,
  MultiTenancyChoiceOptions,
  MultiTenancyModeOptions,
} from './config.options';
// import * as XLSX from 'ts-xlsx';

interface Props extends DataSourcePluginOptionsEditorProps<OCIDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  render() {
    const { options } = this.props;

    const cmdbFileValidator = (cmdbFile: File) => {
      let fName = cmdbFile.name;
      let ext = fName.substr(fName.lastIndexOf('.') + 1);
      if (ext !== 'xlsx') {
        return { code: 'file-invalid-type', message: 'Only excel (.xlsx) is supported' };
      }

      return null;
    };

    const readCMDBExcelFile = (result: string | ArrayBuffer | null) => {
      if (result === null || typeof result === 'string') {
        return;
      }

      let data = new Uint8Array(result);
      let arr: any[] = [];
      for (let i = 0; i !== data.length; ++i) {
        arr[i] = String.fromCharCode(data[i]);
      }

    //   let bstr = arr.join('');
    //   let workbook = XLSX.read(bstr, { type: 'binary' });
    //   let cmdbData: any = {};
    //   for (let ws_name of workbook.SheetNames) {
    //     let ws = workbook.Sheets[ws_name];
    //     cmdbData[ws_name] = XLSX.utils.sheet_to_json(ws, { raw: true });
    //   }

    //   options.jsonData.cmdbFileContent = JSON.stringify(cmdbData);
    };

    let cmdbFileOptions = { maxFiles: 1, multiple: false, validator: cmdbFileValidator };

    return (
      <FieldSet label="Connection Details">
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
        {options.jsonData.authProvider === AuthProviders.OCI_CLI && (
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
              label="Config Profile Name"
              labelWidth={28}
              tooltip="Config profile name, as specified in oci config file. Default value is DEFAULT."
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
            {options.jsonData.enableCMDBUploadFile === true && (
              <>
                <FileDropzone options={cmdbFileOptions} onLoad={readCMDBExcelFile} readAs={'readAsArrayBuffer'} />
              </>
            )}
          </>
        )}
      </FieldSet>
    );
  }
}
import { SelectableValue } from '@grafana/data';

export enum AuthProviders {
  OCI_CLI = 'oci-cli',
  OCI_INSTANCE = 'oci-instance',
}

export enum MultiTenancyChoices {
  YES = 'yes',
  NO = 'no',
}

export enum MultiTenancyModes {
  MULTI_PROFILE = 'multi-profile',
  CROSS_TENANCY_POLICY = 'cross-tenancy-policy',
}

export const AuthProviderOptions = [
  {
    label: 'OCI CLI',
    value: AuthProviders.OCI_CLI,
    description: 'The grafana instance is configured with oci cli',
  },
  {
    label: 'OCI Instance',
    value: AuthProviders.OCI_INSTANCE,
    description: 'The grafana instance is configured in OCI environment',
  },
] as Array<SelectableValue<string>>;

export const MultiTenancyChoiceOptions = [
  {
    label: 'YES',
    value: MultiTenancyChoices.YES,
  },
  {
    label: 'NO',
    value: MultiTenancyChoices.NO,
  },
] as Array<SelectableValue<string>>;

export const MultiTenancyModeOptions = [
  {
    label: 'MULTI PROFILE',
    value: MultiTenancyModes.MULTI_PROFILE,
    description: `Here it is expected user will create multiple profile named 
    with tenancy name in oci configuration file and all tenancies will be in  
    tenancy file with format (<tenancy_name>,<tenancy_ocid>).`,
  },
  {
    label: 'CROSS TENANCY POLICY',
    value: MultiTenancyModes.CROSS_TENANCY_POLICY,
    description: `Here it is expected user will use cross-tenancy IAM policy 
    for a particular user and all tenancies will be in tenancy file with 
    format (<tenancy_name>,<tenancy_ocid>).`,
  },
] as Array<SelectableValue<string>>;

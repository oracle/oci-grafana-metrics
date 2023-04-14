/*
** Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import React from 'react';
import { regions, environments, tenancymodes } from './constants'



interface Profile {
  name: string;
  region?: string;
  userOcid?: string;
  tenancyOcid?: string;
  fingerprint?: string;
  apiKey?: string;
  environments: string;
  tenancyMode: string;  
}

interface Props {
  profiles: Profile[];
  onProfileChange: (key: keyof Profile, value: string, index: number) => void;
  onRemoveProfile: (index: number) => void;
  onAddProfile: () => void;
//   environments: 'OCI' | 'other';
//   tenancyMode: 'single' | 'multi';
  environment: typeof environments[number];
  tenancyMode: typeof tenancymodes[number];
  region?: typeof regions[number];
  onRegionChange?: (region?: typeof regions[number]) => void;  
}

interface RegionSelectorProps {
    value?: string;
    onChange?: (value?: string) => void;
  }

export const OCIConfigCtrl: React.FC<Props> = ({
  profiles,
  onProfileChange,
  onRemoveProfile,
  onAddProfile,
  environment,
  tenancyMode,
}) => {
  const numProfiles = profiles.length;

  return (
    <>
      <div className="gf-form">
        <label className="gf-form-label width-12">Environment</label>
        <select
          className="gf-form-select width-25"
          value={environment}
          onChange={(e) => {
            onProfileChange('region', '', 0); // reset region when switching environmentss
            onProfileChange('tenancyMode', '', 0); // reset tenancy mode when switching environmentss
            onProfileChange('environments', e.target.value as string, 0);
          }}
        >
          <option value="OCI">OCI</option>
          <option value="other">Other</option>
        </select>
      </div>
      {environment === 'OCI' && (
        <div className="gf-form">
          <label className="gf-form-label width-12">Region</label>
          <select
            className="gf-form-select width-25"
            value={profiles[0]?.region || ''}
            onChange={(e) => onProfileChange('region', e.target.value, 0)}
          >
            <option value=""></option>
            {regions.map((region) => (
                <option key={region} value={region}>
                {region}
                </option>
            ))}
          </select>
        </div>
      )}
      {environment === 'other' && (
        <>
          <div className="gf-form">
            <label className="gf-form-label width-12">Tenancy mode</label>
            <select
              className="gf-form-select width-25"
              value={profiles[0]?.tenancyMode || ''}
              onChange={(e) => onProfileChange('tenancyMode', e.target.value, 0)}
            >
              <option value=""></option>
              <option value="single">Single</option>
              <option value="multi">Multi</option>
            </select>
          </div>
          {tenancyMode === 'single' && (
            <div className="gf-form">
              <label className="gf-form-label width-12">Profile Name</label>
              <input
                type="text"
                className="gf-form-input width-25"
                value={profiles[0]?.name || ''}
                onChange={(e) => onProfileChange('name', e.target.value, 0)}
              />
            </div>
          )}
          {tenancyMode === 'multi' &&
            profiles.map((profile, index) => (
              <div key={index}>
                <h4>Profile {index + 1}</h4>
                <div className="gf-form">
                  <label className="gf-form-label width-12">Profile Name</label>
                  <input
                    type="text"
                    className="gf-form-input width-25"
                    value={profile.name || ''}
                    onChange={(e) => onProfileChange('name', e.target.value, index)}
                  />
                  {numProfiles > 1 && (
                    <button className="gf-form-label gf-form-label--btn width-4" onClick={() => onRemoveProfile(index)}>
                      Remove
                    </button>
                  )}
                </div>
                <div className="gf-form">
                  <label className="gf-form-label width-12">User OCID</label>
                  <input
                    type="password"
                    className="gf-form-input width-25"
                    value={profile.userOcid || ''}
                    onChange={(e) => onProfileChange('userOcid', e.target.value, index)}
                  />
                </div>
                <div className="gf-form">
                  <label className="gf-form-label width-12">Tenancy OCID</label>
                  <input
                    type="password"
                    className="gf-form-input width-25"
                    value={profile.tenancyOcid || ''}
                    onChange={(e) => onProfileChange('tenancyOcid', e.target.value, index)}
                  />
                </div>
                <div className="gf-form">
                  <label className="gf-form-label width-12">Fingerprint</label>
                  <input
                    type="password"
                    className="gf-form-input width-25"
                    value={profile.fingerprint || ''}
                    onChange={(e) => onProfileChange('fingerprint', e.target.value, index)}
                  />
                </div>
                <div className="gf-form">
                  <label className="gf-form-label width-12">API Key</label>
                  <input
                    type="password"
                    className="gf-form-input width-25"
                    value={profile.apiKey || ''}
                    onChange={(e) => onProfileChange('apiKey', e.target.value, index)}
                  />
                </div>
                <hr />
              </div>
            ))}
          {tenancyMode === 'multi' && numProfiles < 6 && (
            <button className="gf-form-label gf-form-label--btn width-4" onClick={onAddProfile}>
              Add Profile
            </button>
          )}
        </>
      )}
    </>
  );
};

export default OCIConfigCtrl;

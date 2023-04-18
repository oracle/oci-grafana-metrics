
import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { OCIDataSource } from './datasource';
import { OCIConfig, OCIQuery } from './types';
import { HorizontalGroup, Input, Label } from '@grafana/ui';

type Props = QueryEditorProps<OCIDataSource, OCIQuery, OCIConfig>;

export class QueryEditor extends PureComponent<Props> {
  render() {
    return (
      <HorizontalGroup>
        <Label>Multiplier</Label>
        <Input
          type="number"
          label="Multiplier"
          value={this.props.query.multiplier}
          onChange={(e) => this.props.onChange({ ...this.props.query, multiplier: e.currentTarget.valueAsNumber })}
        />
      </HorizontalGroup>
    );
  }
}
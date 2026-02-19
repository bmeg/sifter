import type { StepCellProps } from '../types';
import { isRecord } from './utils';
import { BaseStepCell } from './BaseStepCell';

export function ObjectValidateStepCell(props: StepCellProps) {
  const config = props.step.config;
  const schema = isRecord(config) && typeof config.schema === 'string' ? config.schema : undefined;

  return <BaseStepCell {...props} secondaryText={schema ? `schema: ${schema}` : undefined} />;
}

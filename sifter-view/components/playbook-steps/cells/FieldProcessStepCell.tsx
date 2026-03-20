import type { StepCellProps } from '../types';
import { isRecord } from './utils';
import { BaseStepCell } from './BaseStepCell';

export function FieldProcessStepCell(props: StepCellProps) {
  const config = props.step.config;
  const field = isRecord(config) && typeof config.field === 'string' ? config.field : undefined;

  return <BaseStepCell {...props} secondaryText={field ? `field: ${field}` : undefined} />;
}

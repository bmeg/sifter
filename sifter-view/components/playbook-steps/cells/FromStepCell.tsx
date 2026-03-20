import type { StepCellProps } from '../types';
import { isRecord } from './utils';
import { BaseStepCell } from './BaseStepCell';

export function FromStepCell(props: StepCellProps) {
  const source = typeof props.step.config === 'string'
    ? props.step.config
    : isRecord(props.step.config) && typeof props.step.config.from === 'string'
      ? props.step.config.from
      : undefined;

  return <BaseStepCell {...props} secondaryText={source ? `source: ${source}` : undefined} />;
}

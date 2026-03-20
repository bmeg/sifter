import type { StepCellProps } from '../types';
import { getStepSummaryValue } from './utils';
import { BaseStepCell } from './BaseStepCell';

export function GenericStepCell(props: StepCellProps) {
  return <BaseStepCell {...props} secondaryText={getStepSummaryValue(props.step.config)} />;
}

import type { StepCellProps } from '../types';
import { isRecord } from './utils';
import { BaseStepCell } from './BaseStepCell';

export function ProjectStepCell(props: StepCellProps) {
  const config = props.step.config;
  const projectionCount = isRecord(config) ? Object.keys(config).length : undefined;

  return <BaseStepCell {...props} secondaryText={projectionCount ? `${projectionCount} mappings` : undefined} />;
}

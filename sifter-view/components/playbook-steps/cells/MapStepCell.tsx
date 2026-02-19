import type { StepCellProps } from '../types';
import { isRecord } from './utils';
import { BaseStepCell } from './BaseStepCell';

export function MapStepCell(props: StepCellProps) {
  const config = props.step.config;
  const engine = isRecord(config) && typeof config.engine === 'string' ? config.engine : undefined;

  return <BaseStepCell {...props} secondaryText={engine ? `engine: ${engine}` : undefined} />;
}

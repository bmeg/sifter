import type { ComponentType } from 'react';

import type { StepCellProps } from './types';
import {
  FieldProcessStepCell,
  FromStepCell,
  GenericStepCell,
  MapStepCell,
  ObjectValidateStepCell,
  ProjectStepCell,
} from './cells';

export type StepCellComponent = ComponentType<StepCellProps>;

export const STEP_OPERATIONS = [
  'from',
  'split',
  'fieldParse',
  'fieldType',
  'objectValidate',
  'filter',
  'clean',
  'debug',
  'regexReplace',
  'project',
  'map',
  'plugin',
  'flatmap',
  'reduce',
  'distinct',
  'fieldProcess',
  'dropNull',
  'lookup',
  'intervalIntersect',
  'hash',
  'accumulate',
  'uuid',
] as const;

type StepOperation = (typeof STEP_OPERATIONS)[number];

const OPERATION_CELL_COMPONENTS: Record<StepOperation, StepCellComponent> = {
  from: FromStepCell,
  split: GenericStepCell,
  fieldParse: GenericStepCell,
  fieldType: GenericStepCell,
  objectValidate: ObjectValidateStepCell,
  filter: GenericStepCell,
  clean: GenericStepCell,
  debug: GenericStepCell,
  regexReplace: GenericStepCell,
  project: ProjectStepCell,
  map: MapStepCell,
  plugin: GenericStepCell,
  flatmap: GenericStepCell,
  reduce: GenericStepCell,
  distinct: GenericStepCell,
  fieldProcess: FieldProcessStepCell,
  dropNull: GenericStepCell,
  lookup: GenericStepCell,
  intervalIntersect: GenericStepCell,
  hash: GenericStepCell,
  accumulate: GenericStepCell,
  uuid: GenericStepCell,
};

export function getStepCellComponent(operation: string): StepCellComponent {
  if ((STEP_OPERATIONS as readonly string[]).includes(operation)) {
    return OPERATION_CELL_COMPONENTS[operation as StepOperation];
  }
  return GenericStepCell;
}

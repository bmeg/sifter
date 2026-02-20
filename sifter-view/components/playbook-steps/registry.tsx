import type { ComponentType } from 'react';

import type { StepCellProps, StepEditorProps } from './types';
import {
  FieldProcessStepCell,
  FromStepCell,
  GenericStepCell,
  MapStepCell,
  ObjectValidateStepCell,
  ProjectStepCell,
} from './cells';
import { FromStepEditor, JsonStepEditor } from './editors';
import { FilterStepEditor, MapStepEditor, ProjectStepEditor } from './editors';

export type StepCellComponent = ComponentType<StepCellProps>;
export type StepEditorComponent = ComponentType<StepEditorProps>;

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

const OPERATION_EDITOR_COMPONENTS: Record<StepOperation, StepEditorComponent> = {
  from: FromStepEditor,
  split: JsonStepEditor,
  fieldParse: JsonStepEditor,
  fieldType: JsonStepEditor,
  objectValidate: JsonStepEditor,
  filter: FilterStepEditor,
  clean: JsonStepEditor,
  debug: JsonStepEditor,
  regexReplace: JsonStepEditor,
  project: ProjectStepEditor,
  map: MapStepEditor,
  plugin: JsonStepEditor,
  flatmap: JsonStepEditor,
  reduce: JsonStepEditor,
  distinct: JsonStepEditor,
  fieldProcess: JsonStepEditor,
  dropNull: JsonStepEditor,
  lookup: JsonStepEditor,
  intervalIntersect: JsonStepEditor,
  hash: JsonStepEditor,
  accumulate: JsonStepEditor,
  uuid: JsonStepEditor,
};

export function getStepCellComponent(operation: string): StepCellComponent {
  if ((STEP_OPERATIONS as readonly string[]).includes(operation)) {
    return OPERATION_CELL_COMPONENTS[operation as StepOperation];
  }
  return GenericStepCell;
}

export function getStepEditorComponent(operation: string): StepEditorComponent {
  if ((STEP_OPERATIONS as readonly string[]).includes(operation)) {
    return OPERATION_EDITOR_COMPONENTS[operation as StepOperation];
  }
  return JsonStepEditor;
}

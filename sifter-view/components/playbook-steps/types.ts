export type PipelineStep = {
  operation: string;
  config: unknown;
};

export type StepCellProps = {
  step: PipelineStep;
  index: number;
  isLast: boolean;
};

export type StepEditorProps = {
  step: PipelineStep;
  onChangeConfig: (nextConfig: unknown) => void;
};

import { useCallback, useMemo } from 'react';
import type { Dispatch, SetStateAction } from 'react';

import type { Playbook } from '@/lib/playbookApi';
import type { PipelineStep } from './playbook-steps/types';

type RawPipelineStep = Record<string, unknown>;

function parsePipelineStep(stepObj: RawPipelineStep): PipelineStep {
  const [operation, config] = Object.entries(stepObj)[0] ?? ['unknown', null];
  return { operation, config };
}

function serializePipelineStep(step: PipelineStep): RawPipelineStep {
  return { [step.operation]: step.config };
}

type UsePlaybookPipelineEditorParams = {
  playbook: Playbook | null;
  setPlaybook: Dispatch<SetStateAction<Playbook | null>>;
  selectedPipelineName: string | null;
  defaultStepOperation: string;
};

export function usePlaybookPipelineEditor({
  playbook,
  setPlaybook,
  selectedPipelineName,
  defaultStepOperation,
}: UsePlaybookPipelineEditorParams) {
  const selectedPipelineSteps = useMemo(() => {
    if (!selectedPipelineName || !playbook) {
      return [] as PipelineStep[];
    }

    return (playbook.pipelines[selectedPipelineName] ?? []).map((stepObj) =>
      parsePipelineStep(stepObj as RawPipelineStep)
    );
  }, [playbook, selectedPipelineName]);

  const updateSelectedPipelineSteps = useCallback(
    (updater: (currentSteps: PipelineStep[]) => PipelineStep[]) => {
      if (!selectedPipelineName) {
        return;
      }

      setPlaybook((currentPlaybook) => {
        if (!currentPlaybook) {
          return currentPlaybook;
        }

        const currentPipelineSteps = (currentPlaybook.pipelines[selectedPipelineName] ?? []).map((stepObj) =>
          parsePipelineStep(stepObj as RawPipelineStep)
        );
        const updatedPipelineSteps = updater(currentPipelineSteps);

        return {
          ...currentPlaybook,
          pipelines: {
            ...currentPlaybook.pipelines,
            [selectedPipelineName]: updatedPipelineSteps.map(serializePipelineStep),
          },
        };
      });
    },
    [selectedPipelineName, setPlaybook]
  );

  const updateStepOperation = useCallback(
    (index: number, operation: string) => {
      updateSelectedPipelineSteps((currentSteps) =>
        currentSteps.map((step, stepIndex) =>
          stepIndex === index ? { ...step, operation } : step
        )
      );
    },
    [updateSelectedPipelineSteps]
  );

  const updateStepConfig = useCallback(
    (index: number, config: unknown) => {
      updateSelectedPipelineSteps((currentSteps) =>
        currentSteps.map((step, stepIndex) =>
          stepIndex === index ? { ...step, config } : step
        )
      );
    },
    [updateSelectedPipelineSteps]
  );

  const reorderSteps = useCallback(
    (fromIndex: number, toIndex: number) => {
      if (fromIndex === toIndex) {
        return;
      }

      updateSelectedPipelineSteps((currentSteps) => {
        if (
          fromIndex < 0 ||
          toIndex < 0 ||
          fromIndex >= currentSteps.length ||
          toIndex >= currentSteps.length
        ) {
          return currentSteps;
        }

        const reorderedSteps = [...currentSteps];
        const [movedStep] = reorderedSteps.splice(fromIndex, 1);
        reorderedSteps.splice(toIndex, 0, movedStep);
        return reorderedSteps;
      });
    },
    [updateSelectedPipelineSteps]
  );

  const removeStep = useCallback(
    (index: number) => {
      updateSelectedPipelineSteps((currentSteps) =>
        currentSteps.filter((_, stepIndex) => stepIndex !== index)
      );
    },
    [updateSelectedPipelineSteps]
  );

  const addStep = useCallback(() => {
    updateSelectedPipelineSteps((currentSteps) => [
      ...currentSteps,
      { operation: defaultStepOperation, config: {} },
    ]);
  }, [defaultStepOperation, updateSelectedPipelineSteps]);

  return {
    selectedPipelineSteps,
    updateStepOperation,
    updateStepConfig,
    reorderSteps,
    removeStep,
    addStep,
  };
}

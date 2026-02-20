// Provides a focused React hook for editing playbook pipeline steps in the UI,
// handling normalization between stored pipeline shapes and editor-friendly data.
import { useCallback, useMemo } from 'react';
import type { Dispatch, SetStateAction } from 'react';

import type { Playbook } from '@/lib/playbookApi';
import type { PipelineStep } from './playbook-steps/types';

type RawPipelineStep = Record<string, unknown>;

// Normalize a raw pipeline step object (single-key record) into a typed shape.
function parsePipelineStep(stepObj: RawPipelineStep): PipelineStep {
  const [operation, config] = Object.entries(stepObj)[0] ?? ['unknown', null];
  return { operation, config };
}

// Convert a typed step back into the raw single-key object stored in the playbook.
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
  // Derive the currently selected pipeline's steps from the playbook.
  // Memoized to avoid unnecessary recalculation when unrelated state changes.
  const selectedPipelineSteps = useMemo(() => {
    if (!selectedPipelineName || !playbook) {
      return [] as PipelineStep[];
    }

    // Decode raw step objects into a normalized shape for UI consumption.
    return (playbook.pipelines[selectedPipelineName] ?? []).map((stepObj) =>
      parsePipelineStep(stepObj as RawPipelineStep)
    );
  }, [playbook, selectedPipelineName]);

  // Centralized updater that replaces the selected pipeline steps immutably.
  // Safeguards against missing selection or playbook state.
  const updateSelectedPipelineSteps = useCallback(
    (updater: (currentSteps: PipelineStep[]) => PipelineStep[]) => {
      if (!selectedPipelineName) {
        return;
      }

      setPlaybook((currentPlaybook) => {
        if (!currentPlaybook) {
          return currentPlaybook;
        }

        // Rehydrate current steps for a deterministic update.
        const currentPipelineSteps = (currentPlaybook.pipelines[selectedPipelineName] ?? []).map((stepObj) =>
          parsePipelineStep(stepObj as RawPipelineStep)
        );
        const updatedPipelineSteps = updater(currentPipelineSteps);

        // Persist updated steps back into the playbook shape.
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

  // Update only the operation of a specific step.
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

  // Update only the config payload of a specific step.
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

  // Move a step within the selected pipeline.
  const reorderSteps = useCallback(
    (fromIndex: number, toIndex: number) => {
      if (fromIndex === toIndex) {
        return;
      }

      updateSelectedPipelineSteps((currentSteps) => {
        // Validate indices to keep the operation a no-op for invalid inputs.
        if (
          fromIndex < 0 ||
          toIndex < 0 ||
          fromIndex >= currentSteps.length ||
          toIndex >= currentSteps.length
        ) {
          return currentSteps;
        }

        // Clone array before in-place mutation for predictable state updates.
        const reorderedSteps = [...currentSteps];
        const [movedStep] = reorderedSteps.splice(fromIndex, 1);
        reorderedSteps.splice(toIndex, 0, movedStep);
        return reorderedSteps;
      });
    },
    [updateSelectedPipelineSteps]
  );

  // Remove a step at the given index.
  const removeStep = useCallback(
    (index: number) => {
      updateSelectedPipelineSteps((currentSteps) =>
        currentSteps.filter((_, stepIndex) => stepIndex !== index)
      );
    },
    [updateSelectedPipelineSteps]
  );

  // Append a new step with the default operation and an empty config.
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

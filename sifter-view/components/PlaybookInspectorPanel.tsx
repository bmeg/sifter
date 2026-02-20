'use client';

import React, { useState } from 'react';
import type { Node } from '@xyflow/react';
import { FaTrashAlt } from 'react-icons/fa';
import { MdDragIndicator } from 'react-icons/md';
import { getStepEditorComponent } from './playbook-steps/registry';
import type { PipelineStep } from './playbook-steps/types';

type PlaybookInspectorPanelProps = {
  selectedNode: Node | null;
  selectedPipelineName: string | null;
  selectedPipelineSteps: PipelineStep[];
  stepOperations: readonly string[];
  onUpdateStepOperation: (index: number, operation: string) => void;
  onUpdateStepConfig: (index: number, config: unknown) => void;
  onReorderSteps: (fromIndex: number, toIndex: number) => void;
  onRemoveStep: (index: number) => void;
  onAddStep: () => void;
};

export default function PlaybookInspectorPanel({
  selectedNode,
  selectedPipelineName,
  selectedPipelineSteps,
  stepOperations,
  onUpdateStepOperation,
  onUpdateStepConfig,
  onReorderSteps,
  onRemoveStep,
  onAddStep,
}: PlaybookInspectorPanelProps) {
  const [draggingStepIndex, setDraggingStepIndex] = useState<number | null>(null);
  const [dragOverStepIndex, setDragOverStepIndex] = useState<number | null>(null);

  return (
    <aside
      style={{
        width: 360,
        borderLeft: '1px solid #ddd',
        background: '#f8fafc',
        padding: 12,
        overflowY: 'auto',
      }}
    >
      <h3 style={{ margin: 0, fontSize: 16, fontWeight: 600 }}>Inspection Panel</h3>
      {!selectedNode && <p style={{ marginTop: 12 }}>Select a node to inspect.</p>}

      {selectedNode && (
        <div style={{ marginTop: 12 }}>
          <div style={{ marginBottom: 8, fontSize: 13, color: '#334155' }}>
            <strong>Node:</strong> {selectedNode.id}
          </div>
          <div style={{ marginBottom: 12, fontSize: 13, color: '#475569' }}>
            <strong>Type:</strong> {selectedNode.type ?? 'default'}
          </div>

          {!selectedPipelineName && (
            <p style={{ fontSize: 13, color: '#475569' }}>
              This node is not a pipeline. Select a pipeline node to edit step operations.
            </p>
          )}

          {selectedPipelineName && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <div style={{ fontSize: 13, color: '#334155' }}>
                <strong>Pipeline:</strong> {selectedPipelineName}
              </div>

              {selectedPipelineSteps.map((step, index) => {
                const operationOptions = stepOperations.includes(step.operation)
                  ? stepOperations
                  : [step.operation, ...stepOperations];
                const isDragOver =
                  dragOverStepIndex === index && draggingStepIndex !== null && draggingStepIndex !== index;
                const StepEditor = getStepEditorComponent(step.operation);

                return (
                  <div
                    key={`${selectedPipelineName}-${index}`}
                    onDragOver={(event) => {
                      event.preventDefault();
                      if (dragOverStepIndex !== index) {
                        setDragOverStepIndex(index);
                      }
                    }}
                    onDrop={(event) => {
                      event.preventDefault();
                      if (draggingStepIndex === null) {
                        return;
                      }
                      onReorderSteps(draggingStepIndex, index);
                      setDraggingStepIndex(null);
                      setDragOverStepIndex(null);
                    }}
                    onDragLeave={(event) => {
                      const relatedTarget = event.relatedTarget as globalThis.Node | null;
                      if (!relatedTarget || !event.currentTarget.contains(relatedTarget)) {
                        setDragOverStepIndex((currentIndex) => (currentIndex === index ? null : currentIndex));
                      }
                    }}
                    style={{
                      border: '1px solid #d1d5db',
                      borderRadius: 8,
                      padding: 8,
                      display: 'flex',
                      flexDirection: 'column',
                      gap: 8,
                      background: isDragOver ? '#e2e8f0' : '#fff',
                    }}
                  >
                    <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                      <div
                        draggable
                        onDragStart={() => {
                          setDraggingStepIndex(index);
                          setDragOverStepIndex(index);
                        }}
                        onDragEnd={() => {
                          setDraggingStepIndex(null);
                          setDragOverStepIndex(null);
                        }}
                        style={{
                          border: '1px solid #cbd5e1',
                          borderRadius: 4,
                          padding: '2px 4px',
                          color: '#000',
                          cursor: 'grab',
                          userSelect: 'none',
                          background: '#fff',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                        }}
                        aria-label="Drag to reorder"
                        title="Drag to reorder"
                      >
                        <MdDragIndicator size={14} aria-hidden="true" />
                      </div>
                      <span style={{ fontSize: 12, color: '#475569' }}>Step {index + 1}</span>
                      <select
                        value={step.operation}
                        onChange={(event) => {
                          onUpdateStepOperation(index, event.target.value);
                        }}
                        style={{ flex: 1, color: '#000' }}
                      >
                        {operationOptions.map((operationName) => (
                          <option key={operationName} value={operationName}>
                            {operationName}
                          </option>
                        ))}
                      </select>
                    </div>

                    <StepEditor
                      step={step}
                      onChangeConfig={(nextConfig) => onUpdateStepConfig(index, nextConfig)}
                    />

                    <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                      <button
                        type="button"
                        aria-label="Remove step"
                        title="Remove step"
                        style={{
                          color: '#000',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          padding: '4px 6px',
                        }}
                        onClick={() => {
                          onRemoveStep(index);
                        }}
                      >
                        <FaTrashAlt size={14} aria-hidden="true" />
                      </button>
                    </div>
                  </div>
                );
              })}

              <button
                type="button"
                style={{ color: '#000' }}
                onClick={onAddStep}
              >
                Add Step
              </button>
            </div>
          )}
        </div>
      )}
    </aside>
  );
}

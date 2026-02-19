'use client';

import React, { memo, useCallback, useEffect, useMemo, useState } from 'react';

import {
  ReactFlow, 
  useNodesState, 
  useEdgesState, 
  addEdge, 
  Background, 
  Controls,
  Handle,
  Position,
  NodeProps,
  OnConnect,
  Connection
} from "@xyflow/react";

// Import the necessary styles
import "@xyflow/react/dist/style.css";

import dagre from 'dagre';
import type { Node, Edge } from '@xyflow/react';
import { getPlaybook, type Playbook } from '@/lib/playbookApi';
import { getStepCellComponent, STEP_OPERATIONS } from './playbook-steps/registry';
import type { PipelineStep } from './playbook-steps/types';


type PipelineNodeData = {
  label: string;
  steps: PipelineStep[];
};

type RawPipelineStep = Record<string, unknown>;

const BASE_NODE_WIDTH = 170;
const BASE_NODE_HEIGHT = 44;
const PIPELINE_NODE_WIDTH = 260;
const PIPELINE_CELL_HEIGHT = 34;
const PIPELINE_HEADER_HEIGHT = 36;

const PipelineStackNode = memo(function PipelineStackNode({ data }: NodeProps<Node>) {
  const typedData = data as PipelineNodeData;
  const steps = typedData?.steps ?? [];

  return (
    <div
      style={{
        position: 'relative',
        width: PIPELINE_NODE_WIDTH,
        border: '1px solid #ccc',
        borderRadius: 8,
        background: '#fff',
        color: '#000',
        overflow: 'hidden',
      }}
    >
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
      <div
        style={{
          height: PIPELINE_HEADER_HEIGHT,
          display: 'flex',
          alignItems: 'center',
          padding: '0 10px',
          borderBottom: '1px solid #ddd',
          fontWeight: 600,
          color: '#000',
        }}
      >
        {typedData.label}
      </div>
      {steps.map((step: PipelineStep, index: number) => {
        const CellComponent = getStepCellComponent(step.operation);
        return (
          <CellComponent
            key={`${typedData.label}-${step.operation}-${index}`}
            step={step}
            index={index}
            isLast={index === steps.length - 1}
          />
        );
      })}
    </div>
  );
});

// -------------------------------------------------------------------
// Build a React‑Flow graph from a Playbook, using Dagre for layout
// -------------------------------------------------------------------

function getOutputSourcePipeline(outputDefinition: Record<string, any>): string | undefined {
  for (const value of Object.values(outputDefinition)) {
    if (value && typeof value === 'object' && typeof value.from === 'string') {
      return value.from;
    }
  }
  return undefined;
}

function buildGraph(pb: Playbook): { nodes: Node[]; edges: Edge[] } {
  // Create a directed Dagre graph
  const g = new dagre.graphlib.Graph({ directed: true });
  g.setGraph({ rankdir: 'LR', nodesep: 60, ranksep: 100 });
  
  g.setDefaultEdgeLabel(function() { return {}; });

  // ---- INPUT NODES ---------------------------------------------------
  Object.keys(pb.inputs).forEach((name) => {
    const id = `input-${name}`;
    g.setNode(id, {
      label: name,
      width: BASE_NODE_WIDTH,
      height: BASE_NODE_HEIGHT,
      nodeType: 'default',
      nodeData: { label: `INPUT · ${name}` },
    });
  });

  // ---- OUTPUT NODES ---------------------------------------------------
  if (pb.outputs) {
    Object.keys(pb.outputs).forEach((name) => {
      const id = `output-${name}`;
      g.setNode(id, {
        label: name,
        width: BASE_NODE_WIDTH,
        height: BASE_NODE_HEIGHT,
        nodeType: 'default',
        nodeData: { label: `OUTPUT · ${name}` },
      });
    });
  }

  // ---- PIPELINE NODES --------------------------------------------------
  Object.entries(pb.pipelines).forEach(([pipelineName, steps]) => {
    const stepData: PipelineStep[] = steps.map((stepObj) => {
      const [operation, config] = Object.entries(stepObj)[0] ?? ['unknown', undefined];
      return { operation, config };
    });
    const nodeHeight = PIPELINE_HEADER_HEIGHT + Math.max(1, stepData.length) * PIPELINE_CELL_HEIGHT;

    g.setNode(pipelineName, {
      label: pipelineName,
      width: PIPELINE_NODE_WIDTH,
      height: nodeHeight,
      nodeType: 'pipeline',
      nodeData: { label: pipelineName, steps: stepData },
    });

    const firstStep = steps[0] as any;
    const fromName = firstStep?.from;
    if (fromName) {
      if (pb.inputs?.[fromName]) {
        g.setEdge(`input-${fromName}`, pipelineName);
      } else if (pb.pipelines?.[fromName]) {
        g.setEdge(fromName, pipelineName);
      }
    }
  });

  if (pb.outputs) {
    Object.entries(pb.outputs).forEach(([outputName, outputDefinition]) => {
      const sourcePipeline = getOutputSourcePipeline(outputDefinition);
      if (sourcePipeline && pb.pipelines?.[sourcePipeline]) {
        g.setEdge(sourcePipeline, `output-${outputName}`);
      }
    });
  }

  // Run the layout algorithm
  dagre.layout(g);

  // Convert Dagre nodes/edges to React‑Flow structures
  const nodes: Node[] = g.nodes().map((id) => {
    const nodeDefinition = g.node(id) as {
      x: number;
      y: number;
      width: number;
      height: number;
      label?: string;
      nodeType?: string;
      nodeData?: Record<string, any>;
    };

    const { x, y, label, width, height, nodeType, nodeData } = nodeDefinition;
    const isInputNode = id.startsWith('input-');
    const isOutputNode = id.startsWith('output-');

    return {
      id,
      position: { x: x - width / 2, y: y - height / 2 },
      data: nodeData ?? { label },
      type: nodeType ?? 'default',
      targetPosition: Position.Left,
      sourcePosition: Position.Right,
      style: nodeType === 'pipeline'
        ? undefined
        : {
            width: BASE_NODE_WIDTH,
            height: BASE_NODE_HEIGHT,
            color: '#000',
            fontWeight: 600,
            borderWidth: 2,
            borderStyle: 'solid',
            borderColor: isInputNode ? '#3b82f6' : isOutputNode ? '#16a34a' : '#999',
            backgroundColor: isInputNode ? '#eff6ff' : isOutputNode ? '#f0fdf4' : '#fff',
          },
    } as Node;
  });

  const edges: Edge[] = g.edges().map((e) => ({
    id: `e-${e.v}-${e.w}`,
    source: e.v,
    target: e.w,
    animated: true,
    style: {
      stroke: '#64748b',
      strokeWidth: 2.5,
    },
  }));

  return { nodes, edges };
}

function parsePipelineStep(stepObj: RawPipelineStep): PipelineStep {
  const [operation, config] = Object.entries(stepObj)[0] ?? ['unknown', null];
  return { operation, config };
}

function serializePipelineStep(step: PipelineStep): RawPipelineStep {
  return { [step.operation]: step.config };
}

// -------------------------------------------------------------------
// React component – PlaybookFlow
// -------------------------------------------------------------------


export default function PlaybookFlow() {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [playbook, setPlaybook] = useState<Playbook | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [configDrafts, setConfigDrafts] = useState<Record<number, string>>({});
  const [configErrors, setConfigErrors] = useState<Record<number, string>>({});
  const [draggingStepIndex, setDraggingStepIndex] = useState<number | null>(null);
  const [dragOverStepIndex, setDragOverStepIndex] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const nodeTypes = useMemo(() => ({ pipeline: PipelineStackNode }), []);

  const selectedNode = useMemo(
    () => nodes.find((node) => node.id === selectedNodeId) ?? null,
    [nodes, selectedNodeId]
  );

  const selectedPipelineName = useMemo(() => {
    if (!selectedNodeId || !playbook) {
      return null;
    }
    return playbook.pipelines[selectedNodeId] ? selectedNodeId : null;
  }, [playbook, selectedNodeId]);

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
    [selectedPipelineName]
  );

  const updateStepAtIndex = useCallback(
    (index: number, updater: (step: PipelineStep) => PipelineStep) => {
      updateSelectedPipelineSteps((currentSteps) =>
        currentSteps.map((step, stepIndex) => (stepIndex === index ? updater(step) : step))
      );
    },
    [updateSelectedPipelineSteps]
  );

  const clearConfigUiState = useCallback(() => {
    setConfigDrafts({});
    setConfigErrors({});
  }, []);

  useEffect(() => {
    let isMounted = true;

    const loadPlaybook = async () => {
      try {
        setLoadError(null);
        setIsLoading(true);
        const loadedPlaybook = await getPlaybook();
        if (!isMounted) {
          return;
        }
        setPlaybook(loadedPlaybook);
      } catch (error) {
        if (!isMounted) {
          return;
        }
        setLoadError(error instanceof Error ? error.message : 'Failed to load playbook');
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    };

    void loadPlaybook();

    return () => {
      isMounted = false;
    };
  }, [setEdges, setNodes]);

  useEffect(() => {
    if (!playbook) {
      return;
    }

    const graph = buildGraph(playbook);
    setNodes(graph.nodes);
    setEdges(graph.edges);
  }, [playbook, setEdges, setNodes]);

  useEffect(() => {
    clearConfigUiState();
  }, [clearConfigUiState, selectedPipelineName]);

  const onConnect: OnConnect = useCallback(
    (connection: Connection) => setEdges((currentEdges) => addEdge(connection, currentEdges)),
    [setEdges]
  );

  const handleConfigDraftChange = useCallback((stepIndex: number, draftValue: string) => {
    setConfigDrafts((currentDrafts) => ({
      ...currentDrafts,
      [stepIndex]: draftValue,
    }));

    try {
      JSON.parse(draftValue);
      setConfigErrors((currentErrors) => {
        const { [stepIndex]: _, ...remainingErrors } = currentErrors;
        return remainingErrors;
      });
    } catch {
      setConfigErrors((currentErrors) => ({
        ...currentErrors,
        [stepIndex]: 'Invalid JSON',
      }));
    }
  }, []);

  const commitConfigDraft = useCallback(
    (stepIndex: number) => {
      const draftValue = configDrafts[stepIndex];
      if (draftValue === undefined) {
        return;
      }

      let parsedConfig: unknown;
      try {
        parsedConfig = JSON.parse(draftValue);
      } catch {
        return;
      }

      updateStepAtIndex(stepIndex, (step) => ({
        ...step,
        config: parsedConfig,
      }));
    },
    [configDrafts, updateStepAtIndex]
  );

  const reorderSelectedPipelineSteps = useCallback(
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

      clearConfigUiState();
    },
    [clearConfigUiState, updateSelectedPipelineSteps]
  );

  if (loadError) {
    return (
      <div style={{ width: '100vw', height: '500px', border: '1px solid #ccc', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        Failed to load playbook: {loadError}
      </div>
    );
  }

  if (isLoading) {
    return (
      <div style={{ width: '100vw', height: '500px', border: '1px solid #ccc', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        Loading playbook...
      </div>
    );
  }

  return (
    <div style={{ width: '100%', height: '90%', border: '1px solid #ccc', display: 'flex' }}>
      <div style={{ flex: 1, minWidth: 0 }}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeClick={(_, node) => setSelectedNodeId(node.id)}
          onPaneClick={() => setSelectedNodeId(null)}
          fitView
        >
          <Background color="#aaa" gap={16} />
          <Controls />
        </ReactFlow>
      </div>
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
                  const currentDraft = configDrafts[index] ?? JSON.stringify(step.config ?? null, null, 2);
                  const operationOptions = (STEP_OPERATIONS as readonly string[]).includes(step.operation)
                    ? STEP_OPERATIONS
                    : [step.operation, ...STEP_OPERATIONS];
                  const isDragOver = dragOverStepIndex === index && draggingStepIndex !== null && draggingStepIndex !== index;

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
                        reorderSelectedPipelineSteps(draggingStepIndex, index);
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
                          <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
                            <circle cx="4" cy="3" r="1" fill="currentColor" />
                            <circle cx="4" cy="7" r="1" fill="currentColor" />
                            <circle cx="4" cy="11" r="1" fill="currentColor" />
                            <circle cx="10" cy="3" r="1" fill="currentColor" />
                            <circle cx="10" cy="7" r="1" fill="currentColor" />
                            <circle cx="10" cy="11" r="1" fill="currentColor" />
                          </svg>
                        </div>
                        <span style={{ fontSize: 12, color: '#475569' }}>Step {index + 1}</span>
                        <select
                          value={step.operation}
                          onChange={(event) => {
                            const nextOperation = event.target.value;
                            updateStepAtIndex(index, (currentStep) => ({
                              ...currentStep,
                              operation: nextOperation,
                            }));
                            clearConfigUiState();
                          }}
                          style={{ flex: 1 }}
                        >
                          {operationOptions.map((operationName) => (
                            <option key={operationName} value={operationName}>
                              {operationName}
                            </option>
                          ))}
                        </select>
                      </div>

                      <textarea
                        value={currentDraft}
                        onChange={(event) => handleConfigDraftChange(index, event.target.value)}
                        onBlur={() => commitConfigDraft(index)}
                        rows={5}
                        style={{ width: '100%', fontFamily: 'monospace', fontSize: 12, resize: 'vertical', color: '#000', background: '#fff' }}
                      />

                      {configErrors[index] && (
                        <div style={{ color: '#b91c1c', fontSize: 12 }}>{configErrors[index]}</div>
                      )}

                      <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                        <button
                          type="button"
                          style={{ color: '#000' }}
                          onClick={() => {
                            updateSelectedPipelineSteps((currentSteps) =>
                              currentSteps.filter((_, stepIndex) => stepIndex !== index)
                            );
                            clearConfigUiState();
                          }}
                        >
                          Remove
                        </button>
                      </div>
                    </div>
                  );
                })}

                <button
                  type="button"
                  style={{ color: '#000' }}
                  onClick={() => {
                    updateSelectedPipelineSteps((currentSteps) => [
                      ...currentSteps,
                      { operation: STEP_OPERATIONS[0], config: {} },
                    ]);
                    clearConfigUiState();
                  }}
                >
                  Add Step
                </button>
              </div>
            )}
          </div>
        )}
      </aside>
    </div>
  );
}
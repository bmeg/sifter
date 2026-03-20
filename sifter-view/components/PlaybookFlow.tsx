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
import { getFiles, getPlaybook, type Playbook, type PlaybookFileNode } from '@/lib/playbookApi';
import { getStepCellComponent, STEP_OPERATIONS } from './playbook-steps/registry';
import PlaybookInspectorPanel from './PlaybookInspectorPanel';
import type { PipelineStep } from './playbook-steps/types';
import { usePlaybookPipelineEditor } from './usePlaybookPipelineEditor';


type PipelineNodeData = {
  label: string;
  steps: PipelineStep[];
};

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

function buildGraph(pb: Playbook | null): { nodes: Node[]; edges: Edge[] } {
  if (!pb || !pb.inputs || !pb.pipelines) {
    return { nodes: [], edges: [] };
  }

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
      const sourcePipeline = getOutputSourcePipeline(outputDefinition as Record<string, any>);
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

function findFirstYamlFile(nodes: PlaybookFileNode[]): string | null {
  for (const node of nodes) {
    if (node.isDir && node.children) {
      const childMatch = findFirstYamlFile(node.children);
      if (childMatch) {
        return childMatch;
      }
      continue;
    }

    if (/\.ya?ml$/i.test(node.path)) {
      return node.path;
    }
  }

  return null;
}

// -------------------------------------------------------------------
// React component – PlaybookFlow
// -------------------------------------------------------------------


export default function PlaybookFlow() {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [playbook, setPlaybook] = useState<Playbook | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
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

  const {
    selectedPipelineSteps,
    updateStepOperation,
    updateStepConfig,
    reorderSteps,
    removeStep,
    addStep,
  } = usePlaybookPipelineEditor({
    playbook,
    setPlaybook,
    selectedPipelineName,
    defaultStepOperation: STEP_OPERATIONS[0],
  });

  useEffect(() => {
    let isMounted = true;

    const loadPlaybook = async () => {
      try {
        setLoadError(null);
        setIsLoading(true);
        const files = await getFiles();
        const playbookPath = findFirstYamlFile(files);
        if (!playbookPath) {
          throw new Error('No YAML playbook files found');
        }

        const loadedPlaybook = await getPlaybook(playbookPath);
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

  const onConnect: OnConnect = useCallback(
    (connection: Connection) => setEdges((currentEdges) => addEdge(connection, currentEdges)),
    [setEdges]
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
    <div style={{ width: '100%', height: '100%', border: '1px solid #ccc', display: 'flex' }}>
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
      <PlaybookInspectorPanel
        selectedNode={selectedNode}
        selectedPipelineName={selectedPipelineName}
        selectedPipelineSteps={selectedPipelineSteps}
        stepOperations={STEP_OPERATIONS}
        onUpdateStepOperation={updateStepOperation}
        onUpdateStepConfig={updateStepConfig}
        onReorderSteps={reorderSteps}
        onRemoveStep={removeStep}
        onAddStep={addStep}
      />
    </div>
  );
}
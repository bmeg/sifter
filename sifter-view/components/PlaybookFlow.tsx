'use client';

import React, { useCallback, useEffect } from 'react';

import {
  ReactFlow, 
  useNodesState, 
  useEdgesState, 
  addEdge, 
  Background, 
  Controls,
  OnConnect,
  Connection
} from "@xyflow/react";

// Import the necessary styles
import "@xyflow/react/dist/style.css";

import dagre from 'dagre';
import type { Node, Edge } from '@xyflow/react';


// -------------------------------------------------------------------
// Types for the Sifter playbook JSON (imported as a static module)
// -------------------------------------------------------------------
interface Playbook {
  class?: string;
  name?: string;
  outdir: string;
  inputs: Record<string, any>;
  outputs?: Record<string, any>;
  pipelines: Record<string, PipelineStep[]>;
  params?: any;
}

type PipelineStep = Record<string, any>; // each step is a single‑key object



var playbook : Playbook = {
  "inputs": {
    "caseData": {
      "json": {
        "path": "{{params.cases}}"
      }
    }
  },
  "name": "gdc",
  "outdir": "output/",
  "outputs": {
    "caseFile": {
      "json": {
        "from": "caseObject",
        "path": "gdc.caseObject.case.json.gz"
      }
    },
    "caseGraph": {
      "graph": {
        "EdgeFix": {
          "gpython": {
            "$ref": "test.py"
          },
          "method": "test"
        },
        "from": "caseObject",
        "path": "gdc.caseGraph",
        "schema": "{{params.schema}}",
        "title": "Case"
      }
    }
  },
  "params": {
    "cases": {
      "default": "../../resources/gdc-case.json.gz",
      "type": "file"
    },
    "schema": {
      "default": "../../resources/schemas",
      "type": "path"
    }
  },
  "pipelines": {
    "caseObject": [
      {
        "from": "caseData"
      },
      {
        "project": {
          "mapping": {
            "experiments": "exp:{{row.project.project_id}}",
            "studies": "{{row.project.project_id}}",
            "type": "case"
          }
        }
      },
      {
        "objectValidate": {
          "schema": "{{params.schema}}",
          "title": "Case"
        }
      }
    ]
  }
}


// -------------------------------------------------------------------
// Build a React‑Flow graph from a Playbook, using Dagre for layout
// -------------------------------------------------------------------

const NODE_WIDTH = 150;
const NODE_HEIGHT = 40;

function buildGraph(pb: Playbook): { nodes: Node[]; edges: Edge[] } {
  // Create a directed Dagre graph
  const g = new dagre.graphlib.Graph({ directed: true });
  //g.setGraph({ rankdir: 'LR', nodesep: 80, ranksep: 120 });
  g.setGraph({});
  
  g.setDefaultEdgeLabel(function() { return {}; });

  // ---- INPUT NODES ---------------------------------------------------
  Object.keys(pb.inputs).forEach((name) => {
    const id = `input-${name}`;
    g.setNode(id, { label: name, width: NODE_WIDTH, height: NODE_HEIGHT });
  });

  // ---- OUTPUT NODES ---------------------------------------------------
  if (pb.outputs) {
    Object.keys(pb.outputs).forEach((name) => {
      const id = `output-${name}`;
      g.setNode(id, { label: name, width: NODE_WIDTH, height: NODE_HEIGHT });
    });
  }

  // ---- PIPELINE STEP NODES & EDGES -----------------------------------
  Object.entries(pb.pipelines).forEach(([pipelineName, steps]) => {
    steps.forEach((stepObj, idx) => {
      const stepKey = Object.keys(stepObj)[0]; // e.g. "from", "fieldProcess", "emit", …
      const nodeId = `${pipelineName}-${idx}`;
      g.setNode(nodeId, { label: stepKey, width: NODE_WIDTH, height: NODE_HEIGHT });

      // Edge from previous step (if any)
      if (idx > 0) {
        const prevId = `${pipelineName}-${idx - 1}`;
        g.setEdge(prevId, nodeId);
      } else {
        // First step – connect to the input referenced by the "from" field
        const fromName = (stepObj as any).from;
        if (fromName) {
          g.setEdge(`input-${fromName}`, nodeId);
        }
      }

      // If this step is an emit and an explicit output exists, connect to it
      if (stepKey === 'emit' && pb.outputs) {
        const emitName = (stepObj as any).emit?.name;
        if (emitName && pb.outputs[emitName]) {
          g.setEdge(nodeId, `output-${emitName}`);
        }
      }
    });
  });

  // Run the layout algorithm
  dagre.layout(g);

  // Convert Dagre nodes/edges to React‑Flow structures
  const nodes: Node[] = g.nodes().map((id) => {
    const { x, y, label } = g.node(id);
    return {
      id,
      position: { x, y },
      data: { label },
      // Optional: give each node a consistent size for the UI
      style: { width: NODE_WIDTH, height: NODE_HEIGHT },
    } as Node;
  });

  const edges: Edge[] = g.edges().map((e) => ({
    id: `e-${e.v}-${e.w}`,
    source: e.v,
    target: e.w,
    animated: true,
  }));

  return { nodes, edges };
}

// -------------------------------------------------------------------
// React component – PlaybookFlow
// -------------------------------------------------------------------


export default function PlaybookFlow() {

  
  var initialGraph = buildGraph(playbook);

  var initialEdges: Edge[] = initialGraph.edges;
  var initialNodes: Node[] = initialGraph.nodes;

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

    const onConnect: OnConnect = useCallback(
        (connection: Connection) => setEdges((edges) => addEdge(connection, edges)),
        [setEdges]
    );

  return (
    <div style={{ width: '100vw', height: '500px', border: '1px solid #ccc' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        fitView
      >
        <Background color="#aaa" gap={16} />
        <Controls />
      </ReactFlow>
    </div>
  );
}
'use client';

import React, { useCallback } from 'react';

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

const initialNodes = [
  { id: '1', position: { x: 0, y: 0 }, data: { label: 'Start Node' } },
  { id: '2', position: { x: 0, y: 100 }, data: { label: 'End Node' } },
];

const initialEdges = [{ id: 'e1-2', source: '1', target: '2', animated: true }];

export default function PlaybookFlow() {
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
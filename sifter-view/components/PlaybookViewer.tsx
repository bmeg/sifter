"use client";

import React, { useEffect, useMemo, useState } from "react";
import { Tree, type TreeNodeData, useTree } from "@mantine/core";
import { FiChevronRight as IconChevronRight } from "react-icons/fi";
import { getFiles, getPlaybook, type PlaybookFileNode } from "@/lib/playbookApi";
import PlaybookFlow from "./PlaybookFlow";

// Types for API responses
type FileTree = PlaybookFileNode[];

export default function PlaybookViewer() {
  const [playbooks, setPlaybooks] = useState<FileTree>([]);
  const tree = useTree();

  const toTreeData = (nodes: FileTree): TreeNodeData[] =>
    nodes.map((node) => ({
      value: node.path,
      label: node.name,
      children: node.children ? toTreeData(node.children) : undefined,
    }));

  const indexNodes = (nodes: FileTree): Record<string, PlaybookFileNode> => {
    const indexed: Record<string, PlaybookFileNode> = {};
    const walk = (items: FileTree) => {
      items.forEach((item) => {
        indexed[item.path] = item;
        if (item.children && item.children.length > 0) {
          walk(item.children);
        }
      });
    };
    walk(nodes);
    return indexed;
  };

  const treeData = useMemo(() => toTreeData(playbooks), [playbooks]);
  const playbookIndex = useMemo(() => indexNodes(playbooks), [playbooks]);

  // Load list of playbooks on mount
  useEffect(() => {
    getFiles()
      .then((data: FileTree) => setPlaybooks(data))
      .catch((err) => console.error(err));
  }, []);

  // Load selected playbook content
  const loadPlaybook = (name: string) => {
    getPlaybook(name)
      .then((playbook) => {
        console.log("Loaded playbook", playbook);
      })
      .catch((err) => console.error(err));
  };

  const handleNodeClick = (value: string) => {
    const node = playbookIndex[value];
    if (!node || node.isDir) {
      return;
    }

    if (/\.ya?ml$/i.test(node.path)) {
      loadPlaybook(node.path);
    }
  };

  return (
    <div className="flex w-[100vw] h-[100vh] backdrop-blur-xl rounded-xl overflow-hidden shadow-2xl bg-gradient-to-br from-[#1e1e2f] to-[#2a2a3d]">
      {/* Sidebar */}
      <aside className="flex-none w-72 bg-white p-5 overflow-y-auto text-slate-900">
        <h1 className="text-lg mb-4 text-slate-700">Files</h1>
        <Tree
          data={treeData}
          tree={tree}
          selectOnClick
          renderNode={({ node, elementProps, expanded, hasChildren }) => {
            const isYamlFile = /\.ya?ml$/i.test(node.value);

            return (
              <div
                {...elementProps}
                className="flex items-center gap-1"
                onClick={(event) => {
                  elementProps.onClick(event);
                  handleNodeClick(node.value);
                }}
              >
                {hasChildren ? (
                  <IconChevronRight
                    size={14}
                    className={`transition-transform duration-150 ${expanded ? "rotate-90" : ""}`}
                  />
                ) : (
                  <span className="inline-block w-[14px]" aria-hidden="true" />
                )}
                <span className={isYamlFile ? "font-semibold" : undefined}>{node.label}</span>
              </div>
            );
          }}
          className="text-sm text-slate-900"
        />
      </aside>

      {/* Content */}
      <main className="flex-1 bg-white/4 p-5 overflow-y-auto">
        <PlaybookFlow/>
      </main>
    </div>
  );
}

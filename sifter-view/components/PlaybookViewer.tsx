import React, { useEffect, useState } from "react";
import hljs from "highlight.js";
import "highlight.js/styles/github-dark.css";

// Types for API responses
type PlaybookList = string[];

type PlaybookViewerProps = {};

export default function PlaybookViewer(_: PlaybookViewerProps) {
  const [playbooks, setPlaybooks] = useState<PlaybookList>([]);
  const [selected, setSelected] = useState<string | null>(null);
  const [content, setContent] = useState<string>("");

  // Load list of playbooks on mount
  useEffect(() => {
    fetch("/api/playbooks")
      .then((res) => {
        if (!res.ok) throw new Error("Failed to fetch playbook list");
        return res.json();
      })
      .then((data: PlaybookList) => setPlaybooks(data))
      .catch((err) => console.error(err));
  }, []);

  // Load selected playbook content
  const loadPlaybook = (name: string) => {
    setSelected(name);
    fetch(`/api/playbook?name=${encodeURIComponent(name)}`)
      .then((res) => {
        if (!res.ok) throw new Error("Playbook not found");
        return res.text();
      })
      .then((txt) => setContent(txt))
      .catch((err) => console.error(err));
  };

  // Highlight content after it changes
  useEffect(() => {
    if (content) {
      // Using highlight.js to highlight the content as YAML
      const block = document.getElementById("playbook-content");
      if (block) {
        hljs.highlightElement(block);
      }
    }
  }, [content]);

  return (
    <div className="flex w-[90vw] h-[80vh] backdrop-blur-xl rounded-xl overflow-hidden shadow-2xl bg-gradient-to-br from-[#1e1e2f] to-[#2a2a3d]">
      {/* Sidebar */}
      <aside className="flex-none w-64 bg-white/5 p-5 overflow-y-auto">
        <h1 className="text-lg mb-4 text-[#4f9bff]">Sifter Playbooks</h1>
        <ul>
          {playbooks.map((name) => (
            <li
              key={name}
              onClick={() => loadPlaybook(name)}
              className={`my-2 cursor-pointer px-2 py-1 rounded-md transition-colors hover:bg-white/10 hover:translate-x-1 ${
                selected === name ? "bg-white/10" : ""
              }`}
            >
              {name}
            </li>
          ))}
        </ul>
      </aside>

      {/* Content */}
      <main className="flex-1 bg-white/4 p-5 overflow-y-auto">
        <pre className="h-full overflow-auto m-0">
          <code id="playbook-content" className="language-yaml">
            {content || "Select a playbook to view its content"}
          </code>
        </pre>
      </main>
    </div>
  );
}

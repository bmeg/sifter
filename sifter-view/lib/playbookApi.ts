type PipelineStep = Record<string, unknown>;

export interface PlaybookFileNode {
  name: string;
  path: string;
  isDir: boolean;
  children?: PlaybookFileNode[];
}

export interface Playbook {
  class?: string;
  name?: string;
  outdir?: string;
  inputs: Record<string, unknown>;
  outputs?: Record<string, unknown>;
  pipelines: Record<string, PipelineStep[]>;
  params?: unknown;
}

export async function getFiles(): Promise<PlaybookFileNode[]> {
  const response = await fetch('/api/files');
  if (!response.ok) {
    throw new Error(`Failed to fetch files: ${response.statusText}`);
  }
  return response.json();
}

export async function getPlaybook(name: string): Promise<Playbook> {
  console.log("Fetching playbook", name);
  const response = await fetch(`/api/playbook?name=${encodeURIComponent(name)}&format=json`);
  if (!response.ok) {
    throw new Error(`Failed to fetch playbook: ${response.statusText}`);
  }
  return response.json();
}

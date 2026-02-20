import React, { useEffect, useMemo, useState } from 'react';

import type { StepEditorProps } from '../types';

type StringMap = Record<string, string>;

type AnyMap = Record<string, unknown>;

function asObject(value: unknown): Record<string, unknown> {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

function parseJsonObject<T extends Record<string, unknown>>(value: string): T | null {
  try {
    const parsed = JSON.parse(value);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as T;
    }
  } catch {
    return null;
  }
  return null;
}

export function ProjectStepEditor({ step, onChangeConfig }: StepEditorProps) {
  const configObject = useMemo(() => asObject(step.config), [step.config]);
  const mappingValue = useMemo(
    () => JSON.stringify(asObject(configObject.mapping), null, 2),
    [configObject.mapping]
  );
  const renameValue = useMemo(
    () => JSON.stringify(asObject(configObject.rename), null, 2),
    [configObject.rename]
  );

  const [mappingDraft, setMappingDraft] = useState(mappingValue);
  const [renameDraft, setRenameDraft] = useState(renameValue);
  const [mappingError, setMappingError] = useState<string | null>(null);
  const [renameError, setRenameError] = useState<string | null>(null);

  useEffect(() => {
    setMappingDraft(mappingValue);
    setMappingError(null);
  }, [mappingValue]);

  useEffect(() => {
    setRenameDraft(renameValue);
    setRenameError(null);
  }, [renameValue]);

  const commit = (nextMappingRaw: string, nextRenameRaw: string) => {
    const parsedMapping = parseJsonObject<AnyMap>(nextMappingRaw);
    const parsedRenameRaw = parseJsonObject<Record<string, unknown>>(nextRenameRaw);

    if (!parsedMapping) {
      setMappingError('Mapping must be a JSON object');
      return;
    }

    if (!parsedRenameRaw) {
      setRenameError('Rename must be a JSON object');
      return;
    }

    const parsedRename: StringMap = {};
    for (const [key, value] of Object.entries(parsedRenameRaw)) {
      parsedRename[key] = String(value ?? '');
    }

    setMappingError(null);
    setRenameError(null);

    onChangeConfig({
      ...configObject,
      mapping: parsedMapping,
      rename: parsedRename,
    });
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div>
        <label style={{ fontSize: 12, color: '#334155' }}>Mapping (JSON object)</label>
        <textarea
          value={mappingDraft}
          onChange={(event) => setMappingDraft(event.target.value)}
          onBlur={() => commit(mappingDraft, renameDraft)}
          rows={5}
          style={{
            width: '100%',
            fontFamily: 'monospace',
            fontSize: 12,
            resize: 'vertical',
            color: '#000',
            background: '#fff',
          }}
        />
        {mappingError && <div style={{ color: '#b91c1c', fontSize: 12 }}>{mappingError}</div>}
      </div>

      <div>
        <label style={{ fontSize: 12, color: '#334155' }}>Rename (JSON object)</label>
        <textarea
          value={renameDraft}
          onChange={(event) => setRenameDraft(event.target.value)}
          onBlur={() => commit(mappingDraft, renameDraft)}
          rows={4}
          style={{
            width: '100%',
            fontFamily: 'monospace',
            fontSize: 12,
            resize: 'vertical',
            color: '#000',
            background: '#fff',
          }}
        />
        {renameError && <div style={{ color: '#b91c1c', fontSize: 12 }}>{renameError}</div>}
      </div>
    </div>
  );
}

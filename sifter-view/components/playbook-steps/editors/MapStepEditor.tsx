import React, { useEffect, useMemo, useState } from 'react';

import type { StepEditorProps } from '../types';

type ScriptKind = 'python' | 'gpython';

function asObject(value: unknown): Record<string, unknown> {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

function toDisplayString(value: unknown): string {
  if (typeof value === 'string') {
    return value;
  }
  if (value === undefined || value === null) {
    return '';
  }
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

export function MapStepEditor({ step, onChangeConfig }: StepEditorProps) {
  const configObject = useMemo(() => asObject(step.config), [step.config]);
  const initialMethod = typeof configObject.method === 'string' ? configObject.method : '';
  const hasPython = typeof configObject.python === 'string' && configObject.python.length > 0;
  const initialScriptKind: ScriptKind = hasPython ? 'python' : 'gpython';
  const initialScript = toDisplayString(configObject[initialScriptKind]);

  const [method, setMethod] = useState(initialMethod);
  const [scriptKind, setScriptKind] = useState<ScriptKind>(initialScriptKind);
  const [script, setScript] = useState(initialScript);

  useEffect(() => {
    setMethod(initialMethod);
    setScriptKind(initialScriptKind);
    setScript(initialScript);
  }, [initialMethod, initialScript, initialScriptKind]);

  const commit = (next: { methodValue: string; kind: ScriptKind; scriptValue: string }) => {
    const nextConfig: Record<string, unknown> = {
      ...configObject,
      method: next.methodValue,
      python: undefined,
      gpython: undefined,
    };

    if (next.scriptValue.trim() !== '') {
      nextConfig[next.kind] = next.scriptValue;
    }

    onChangeConfig(nextConfig);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ display: 'flex', gap: 8 }}>
        <div style={{ flex: 1 }}>
          <label style={{ fontSize: 12, color: '#334155' }}>Method</label>
          <input
            type="text"
            value={method}
            onChange={(event) => setMethod(event.target.value)}
            onBlur={() => commit({ methodValue: method, kind: scriptKind, scriptValue: script })}
            style={{
              width: '100%',
              color: '#000',
              background: '#fff',
              border: '1px solid #cbd5e1',
              borderRadius: 4,
              padding: '6px 8px',
            }}
          />
        </div>
        <div style={{ width: 110 }}>
          <label style={{ fontSize: 12, color: '#334155' }}>Engine</label>
          <select
            value={scriptKind}
            onChange={(event) => {
              const nextKind = event.target.value as ScriptKind;
              setScriptKind(nextKind);
              commit({ methodValue: method, kind: nextKind, scriptValue: script });
            }}
            style={{ width: '100%', color: '#000' }}
          >
            <option value="python">python</option>
            <option value="gpython">gpython</option>
          </select>
        </div>
      </div>

      <div>
        <label style={{ fontSize: 12, color: '#334155' }}>
          {scriptKind === 'python' ? 'Python Code' : 'GPython Code'}
        </label>
        <textarea
          value={script}
          onChange={(event) => setScript(event.target.value)}
          onBlur={() => commit({ methodValue: method, kind: scriptKind, scriptValue: script })}
          rows={6}
          style={{
            width: '100%',
            fontFamily: 'monospace',
            fontSize: 12,
            resize: 'vertical',
            color: '#000',
            background: '#fff',
          }}
        />
      </div>
    </div>
  );
}

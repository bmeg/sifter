import React, { useEffect, useMemo, useState } from 'react';

import type { StepEditorProps } from '../types';

type FilterCheck = '' | 'exists' | 'hasValue' | 'not';
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

export function FilterStepEditor({ step, onChangeConfig }: StepEditorProps) {
  const configObject = useMemo(() => asObject(step.config), [step.config]);

  const initialField = typeof configObject.field === 'string' ? configObject.field : '';
  const initialValue = typeof configObject.value === 'string' ? configObject.value : '';
  const initialMatch = typeof configObject.match === 'string' ? configObject.match : '';
  const initialCheck =
    configObject.check === 'exists' || configObject.check === 'hasValue' || configObject.check === 'not'
      ? configObject.check
      : '';
  const initialMethod = typeof configObject.method === 'string' ? configObject.method : '';
  const hasPython = typeof configObject.python === 'string' && configObject.python.length > 0;
  const initialScriptKind: ScriptKind = hasPython ? 'python' : 'gpython';
  const initialScript = toDisplayString(configObject[initialScriptKind]);

  const [field, setField] = useState(initialField);
  const [value, setValue] = useState(initialValue);
  const [match, setMatch] = useState(initialMatch);
  const [check, setCheck] = useState<FilterCheck>(initialCheck);
  const [method, setMethod] = useState(initialMethod);
  const [scriptKind, setScriptKind] = useState<ScriptKind>(initialScriptKind);
  const [script, setScript] = useState(initialScript);

  useEffect(() => {
    setField(initialField);
    setValue(initialValue);
    setMatch(initialMatch);
    setCheck(initialCheck);
    setMethod(initialMethod);
    setScriptKind(initialScriptKind);
    setScript(initialScript);
  }, [initialField, initialValue, initialMatch, initialCheck, initialMethod, initialScriptKind, initialScript]);

  const commit = (next: {
    fieldValue: string;
    valueValue: string;
    matchValue: string;
    checkValue: FilterCheck;
    methodValue: string;
    scriptEngine: ScriptKind;
    scriptValue: string;
  }) => {
    const nextConfig: Record<string, unknown> = {
      ...configObject,
      field: next.fieldValue,
      value: next.valueValue,
      match: next.matchValue,
      check: next.checkValue,
      method: next.methodValue,
      python: undefined,
      gpython: undefined,
    };

    if (next.scriptValue.trim() !== '') {
      nextConfig[next.scriptEngine] = next.scriptValue;
    }

    onChangeConfig(nextConfig);
  };

  const commitCurrent = () => {
    commit({
      fieldValue: field,
      valueValue: value,
      matchValue: match,
      checkValue: check,
      methodValue: method,
      scriptEngine: scriptKind,
      scriptValue: script,
    });
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ display: 'flex', gap: 8 }}>
        <div style={{ flex: 1 }}>
          <label style={{ fontSize: 12, color: '#334155' }}>Field</label>
          <input
            type="text"
            value={field}
            onChange={(event) => setField(event.target.value)}
            onBlur={commitCurrent}
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
          <label style={{ fontSize: 12, color: '#334155' }}>Check</label>
          <select
            value={check}
            onChange={(event) => {
              const nextCheck = event.target.value as FilterCheck;
              setCheck(nextCheck);
              commit({
                fieldValue: field,
                valueValue: value,
                matchValue: match,
                checkValue: nextCheck,
                methodValue: method,
                scriptEngine: scriptKind,
                scriptValue: script,
              });
            }}
            style={{ width: '100%', color: '#000' }}
          >
            <option value="">(none)</option>
            <option value="exists">exists</option>
            <option value="hasValue">hasValue</option>
            <option value="not">not</option>
          </select>
        </div>
      </div>

      <div style={{ display: 'flex', gap: 8 }}>
        <div style={{ flex: 1 }}>
          <label style={{ fontSize: 12, color: '#334155' }}>Value Template</label>
          <input
            type="text"
            value={value}
            onChange={(event) => setValue(event.target.value)}
            onBlur={commitCurrent}
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
        <div style={{ flex: 1 }}>
          <label style={{ fontSize: 12, color: '#334155' }}>Match</label>
          <input
            type="text"
            value={match}
            onChange={(event) => setMatch(event.target.value)}
            onBlur={commitCurrent}
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
      </div>

      <div style={{ display: 'flex', gap: 8 }}>
        <div style={{ flex: 1 }}>
          <label style={{ fontSize: 12, color: '#334155' }}>Method</label>
          <input
            type="text"
            value={method}
            onChange={(event) => setMethod(event.target.value)}
            onBlur={commitCurrent}
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
              commit({
                fieldValue: field,
                valueValue: value,
                matchValue: match,
                checkValue: check,
                methodValue: method,
                scriptEngine: nextKind,
                scriptValue: script,
              });
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
          onBlur={commitCurrent}
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
      </div>
    </div>
  );
}

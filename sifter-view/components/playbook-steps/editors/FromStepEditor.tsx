import React, { useEffect, useState } from 'react';

import type { StepEditorProps } from '../types';

export function FromStepEditor({ step, onChangeConfig }: StepEditorProps) {
  const fromValue = typeof step.config === 'string'
    ? step.config
    : typeof step.config === 'object' && step.config !== null && 'from' in step.config
      ? String((step.config as { from?: unknown }).from ?? '')
      : '';

  const [value, setValue] = useState(fromValue);

  useEffect(() => {
    setValue(fromValue);
  }, [fromValue]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <label style={{ fontSize: 12, color: '#334155' }}>Input Source</label>
      <input
        type="text"
        value={value}
        onChange={(event) => setValue(event.target.value)}
        onBlur={() => onChangeConfig(value)}
        placeholder="input or pipeline name"
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
  );
}

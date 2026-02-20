import React, { useEffect, useMemo, useState } from 'react';

import type { StepEditorProps } from '../types';

export function JsonStepEditor({ step, onChangeConfig }: StepEditorProps) {
  const initialDraft = useMemo(() => JSON.stringify(step.config ?? null, null, 2), [step.config]);
  const [draftValue, setDraftValue] = useState(initialDraft);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setDraftValue(initialDraft);
    setError(null);
  }, [initialDraft]);

  return (
    <div>
      <textarea
        value={draftValue}
        onChange={(event) => {
          const nextValue = event.target.value;
          setDraftValue(nextValue);
          try {
            JSON.parse(nextValue);
            setError(null);
          } catch {
            setError('Invalid JSON');
          }
        }}
        onBlur={() => {
          try {
            const parsedConfig = JSON.parse(draftValue);
            onChangeConfig(parsedConfig);
            setError(null);
          } catch {
            setError('Invalid JSON');
          }
        }}
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
      {error && <div style={{ color: '#b91c1c', fontSize: 12, marginTop: 4 }}>{error}</div>}
    </div>
  );
}

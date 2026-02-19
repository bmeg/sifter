import type { StepCellProps } from '../types';
import { toStepLabel } from './utils';

const PIPELINE_CELL_HEIGHT = 34;

type BaseStepCellProps = StepCellProps & {
  secondaryText?: string;
};

export function BaseStepCell({ step, index, isLast, secondaryText }: BaseStepCellProps) {
  return (
    <div
      style={{
        minHeight: PIPELINE_CELL_HEIGHT,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '6px 10px',
        borderBottom: isLast ? 'none' : '1px solid #eee',
        fontSize: 12,
        color: '#000',
        gap: 10,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, minWidth: 0 }}>
        <span style={{ opacity: 0.7 }}>{index + 1}.</span>
        <span style={{ fontWeight: 500, whiteSpace: 'nowrap' }}>{step.operation}</span>
      </div>
      {secondaryText ? (
        <span
          style={{
            color: '#475569',
            fontSize: 11,
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            maxWidth: '58%',
          }}
        >
          {secondaryText}
        </span>
      ) : null}
    </div>
  );
}

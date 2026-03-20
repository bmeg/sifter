export function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

export function toStepLabel(operation: string): string {
  return operation
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/^./, (x) => x.toUpperCase());
}

export function getStepSummaryValue(config: unknown): string | undefined {
  if (typeof config === 'string' || typeof config === 'number' || typeof config === 'boolean') {
    return String(config);
  }
  if (Array.isArray(config)) {
    return `${config.length} items`;
  }
  if (isRecord(config)) {
    return `${Object.keys(config).length} fields`;
  }
  return undefined;
}
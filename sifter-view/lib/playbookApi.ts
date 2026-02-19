type PipelineStep = Record<string, any>;

export interface Playbook {
  class?: string;
  name?: string;
  outdir?: string;
  inputs: Record<string, any>;
  outputs?: Record<string, any>;
  pipelines: Record<string, PipelineStep[]>;
  params?: any;
}

const STUB_PLAYBOOK: Playbook = {
  class: 'sifter',
  inputs: {
    caseData: {
      jsonLoad: {
        input: '{{params.cases}}',
      },
    },
  },
  name: 'gdc',
  outputs: {
    aliquot: {
      json: {
        from: 'aliquotObject',
        path: 'aliquot.json',
      },
    },
    case: {
      json: {
        from: 'caseObject',
        path: 'case.json',
      },
    },
    project: {
      json: {
        from: 'projectObject',
        path: 'project.json',
      },
    },
    sample: {
      json: {
        from: 'sampleObject',
        path: 'sample.json',
      },
    },
    table: {
      csv: {
        from: 'aliquotAlias',
        path: 'aliquot.table.csv',
      },
    },
  },
  params: {
    cases: {
      default: 'cases.json',
      type: 'path',
    },
    schema: {
      default: 'schema',
      type: 'path',
    },
  },
  pipelines: {
    aliquotAlias: [
      {
        from: 'aliquotData',
      },
      {
        clean: {
          fields: ['aliquot_id', 'submitter_id', 'project_id'],
        },
      },
    ],
    aliquotData: [
      {
        from: 'sampleData',
      },
      {
        fieldProcess: {
          field: 'portions',
          mapping: {
            project_id: '{{row.project_id}}',
            sample: '{{row.sample_id}}',
          },
        },
      },
      {
        fieldProcess: {
          field: 'analytes',
          mapping: {
            project_id: '{{row.project_id}}',
            sample: '{{row.sample}}',
          },
        },
      },
      {
        fieldProcess: {
          field: 'aliquots',
          mapping: {
            project_id: '{{row.project_id}}',
            sample: '{{row.sample}}',
          },
        },
      },
      {
        project: {
          mapping: {
            id: '{{row.aliquot_id}}',
            type: 'aliquot',
          },
        },
      },
    ],
    aliquotObject: [
      {
        from: 'aliquotData',
      },
      {
        project: {
          mapping: {
            sample: [{ id: '{{row.sample}}' }],
          },
        },
      },
      {
        objectValidate: {
          schema: '{{params.schema}}',
          title: 'Aliquot',
        },
      },
    ],
    caseObject: [
      {
        from: 'caseData',
      },
      {
        project: {
          mapping: {
            experiments: 'exp:{{row.project.project_id}}',
            project_id: '{{row.project.project_id}}',
            projects: [{ id: 'project/{{row.project.project_id}}' }],
            studies: '{{row.project.project_id}}',
            type: 'case',
          },
        },
      },
      {
        map: {
          gpython:
            '\ndef fix(x):\n  samples = []\n  for s in x.get(\'samples\', []):\n    samples.append({"id":s["sample_id"]})\n  x[\'samples\'] = samples\n  return x\n',
          method: 'fix',
        },
      },
      {
        objectValidate: {
          schema: '{{params.schema}}',
          title: 'Case',
        },
      },
    ],
    projectObject: [
      {
        from: 'caseObject',
      },
      {
        distinct: {
          value: '{{row.project_id}}',
        },
      },
      {
        project: {
          mapping: {
            id: '{{row.project_id}}',
            submitter_id: '{{row.project_id}}',
            type: 'project',
          },
        },
      },
      {
        clean: {
          fields: ['id', 'project_id', 'submitter_id', 'type'],
        },
      },
      {
        objectValidate: {
          schema: '{{params.schema}}',
          title: 'Project',
        },
      },
    ],
    sampleData: [
      {
        from: 'caseData',
      },
      {
        fieldProcess: {
          field: 'samples',
          mapping: {
            case: '{{row.id}}',
            project_id: '{{row.project.project_id}}',
          },
        },
      },
    ],
    sampleObject: [
      {
        from: 'sampleData',
      },
      {
        project: {
          mapping: {
            id: '{{row.sample_id}}',
            type: 'sample',
          },
        },
      },
      {
        project: {
          mapping: {
            case: [{ id: '{{row.case}}' }],
          },
        },
      },
      {
        objectValidate: {
          schema: '{{params.schema}}',
          title: 'Sample',
        },
      },
    ],
  },
};

const STUB_PLAYBOOKS: Record<string, Playbook> = {
  gdc: STUB_PLAYBOOK,
  demo: {
    ...STUB_PLAYBOOK,
    name: 'demo',
  },
};

const DEFAULT_PLAYBOOK_NAME = 'gdc';
const DEFAULT_STUB_LATENCY_MS = 250;

function getStubLatencyMs(): number {
  // Configure stub delay with NEXT_PUBLIC_PLAYBOOK_STUB_LATENCY_MS (defaults to 250ms).
  const rawValue = process.env.NEXT_PUBLIC_PLAYBOOK_STUB_LATENCY_MS;
  if (!rawValue) {
    return DEFAULT_STUB_LATENCY_MS;
  }

  const parsedValue = Number(rawValue);
  if (!Number.isFinite(parsedValue) || parsedValue < 0) {
    return DEFAULT_STUB_LATENCY_MS;
  }

  return parsedValue;
}

async function sleep(ms: number): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, ms));
}

export async function getPlaybooks(): Promise<string[]> {
  await sleep(getStubLatencyMs());
  return Object.keys(STUB_PLAYBOOKS);
}

export async function getPlaybook(name: string = DEFAULT_PLAYBOOK_NAME): Promise<Playbook> {
  await sleep(getStubLatencyMs());
  const playbook = STUB_PLAYBOOKS[name];
  if (!playbook) {
    throw new Error(`Playbook not found: ${name}`);
  }

  return structuredClone(playbook);
}

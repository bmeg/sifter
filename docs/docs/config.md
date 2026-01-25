---
title: Paramaters
---

## Paramaters Variables

Playbooks can be parameterized. They are defined in the `params` section of the playbook YAML file.

### Configuration Syntax
```yaml
params:
  variableName:
    type: File # one of: File, Path, String, Number
    default: "path/to/default"
```

### Supported Types
- `File`: Represents a file path
- `Dir`: Represents a directory path

### Example Configuration
```yaml
params:
  inputDir:
    type: Dir
    default: "/data/input"
  outputDir:
    type: Dir
    default: "/data/output"
  schemaFile:
    type: File
    default: "/config/schema.json"
```


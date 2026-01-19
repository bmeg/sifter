---
title: Configuration Reference
---

## Configuration Variables

Configuration variables allow playbooks to be parameterized. They are defined in the `config` section of the playbook YAML file.

### Configuration Syntax
```yaml
config:
  variableName:
    type: File # or Dir
    default: "path/to/default"
```

### Supported Types
- `File`: Represents a file path
- `Dir`: Represents a directory path

### Example Configuration
```yaml
config:
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

### Best Practices
1. Use descriptive names for configuration variables
2. Provide reasonable default values
3. Document all configuration variables in your playbook's documentation section
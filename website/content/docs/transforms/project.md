---
title: project
menu:
  main:
    parent: transforms
    weight: 100
---

# project

Populate row with templated values


# parameters

| name | Type | Description |
| --- | --- | --- |
| mapping | map[string]any | New fields to be generated from template |
| rename | map[string]string | Rename field (no template engine) |


# Example

```yaml
    - project:
        mapping:
          type: sample
          id: "{{row.sample_id}}"
```
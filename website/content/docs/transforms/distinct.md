---
title: distinct
menu:
  main:
    parent: transforms
    weight: 100
---

# distinct
Using templated value, allow only the first record for each distinct key

## Parameters

| name | Type | Description |
| --- | --- | --- |
| value | string | Key used for distinct value |

## Example

```yaml
    - distinct:
        value: "{{row.key}}"
```
---
title: flatMap
menu:
  main:
    parent: transforms
    weight: 15
---

# flatMap

Flatten an array field into separate messages, each containing a single element of the array.

## Parameters

| Parameter | Type   | Description |
|-----------|--------|------------|
| `field`   | string | Path to the array field to be flattened (e.g., `{{row.samples}}`). |
| `dest`    | string | Optional name of the field to store the flattened element (defaults to the same field name). |
| `keep`    | bool   | If `true`, keep the original array alongside the flattened messages. |

## Example

```yaml
- flatMap:
    field: "{{row.samples}}"
    dest: sample
```

Given an input message:

```json
{ "id": "P001", "samples": ["S1", "S2", "S3"] }
```

The step emits three messages:

```json
{ "id": "P001", "sample": "S1" }
{ "id": "P001", "sample": "S2" }
{ "id": "P001", "sample": "S3" }
```

## See also

- [filter](filter.md) – conditionally emit messages.
- [map](map.md) – apply a function to each flattened message.

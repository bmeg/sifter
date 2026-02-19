---
title: uuid
menu:
  main:
    parent: transforms
    weight: 100
---

# uuid

Generate a UUID for a field.

## Parameters

| Name | Type | Description |
| --- | --- | --- |
| field | string | Destination field name for the UUID |
| value | string | Seed value used to generate the UUID |
| namespace | string | UUID namespace (optional) |

## Example

```yaml
    - uuid:
        field: id
        value: "{{row.name}}"
```
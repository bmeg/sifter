---
title: fieldParse
menu:
  main:
    parent: transforms
    weight: 100
---

# fieldParse

Parse a string field (e.g. `key1=val1;key2=val2`) into individual keys.

## Parameters

| Name | Type | Description |
| --- | --- | --- |
| field | string | The field containing the string to be parsed |
| sep | string | Separator character used to split the string |

## Example

```yaml
    - fieldParse:
        field: attributes
        sep: ";"
```

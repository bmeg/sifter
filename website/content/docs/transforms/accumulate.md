---
title: accumulate
menu:
  main:
    parent: transforms
    weight: 100
---

# accumulate

Gather sequential rows into a single record, based on matching a field

## Parameters

| name | Type | Description |
| --- | --- | --- |
| field | string (field path) | Field used to match rows | 
| dest | string | field to store accumulated records |

## Example

```
  - accumulate:
      field: model_id
      dest: rows   
```

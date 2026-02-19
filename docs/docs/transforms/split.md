---
title: split
menu:
  main:
    parent: transforms
    weight: 100
---

# split

Split a field using string `sep`
## Parameters

| name | Type | Description |
| --- | --- | --- |
| field | string | Field to the split |
| sep | string | String to use for splitting |

## Example

```yaml
    - split:
        field: methods
        sep: ";"
```
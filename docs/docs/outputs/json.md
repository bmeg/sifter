---
title: json
menu:
  main:
    parent: transforms
    weight: 100
---

# Output: json

Send data to output file. The naming of the file is `outdir`/`path`

## Parameters

| name | Type | Description |
| --- | --- | --- |
| path | string | Path to output file |

## example

```yaml
output:
  outfile:
    json: 
      path: protein_compound_association.ndjson
```
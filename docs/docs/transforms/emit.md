---
title: emit
menu:
  main:
    parent: transforms
    weight: 100
---

# emit

Send data to output file. The naming of the file is `outdir`/`script name`.`pipeline name`.`emit name`.json.gz

## Parameters

| name | Type | Description |
| --- | --- | --- |
| name | string | Name of emit value |

## example

```yaml
    - emit:
        name: protein_compound_association
```
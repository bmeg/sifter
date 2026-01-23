---
title: sqldump
menu:
  main:
    parent: inputs
    weight: 100
---

# sqlDump
Scan file produced produced from sqldump. 

## Parameters

| Name | Type | Description |
|-------|---|--------|
| path | string | Path to the SQL dump file | 
| tables | []string | Names of tables to read out |

## Example

```yaml
inputs:
  database:
    sqldumpLoad:
      path: "{{params.sql}}"
      tables:
        - cells
        - cell_tissues
        - dose_responses
        - drugs
        - drug_annots
        - experiments
        - profiles
```

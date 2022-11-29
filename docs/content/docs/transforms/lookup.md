---
title: lookup
menu:
  main:
    parent: transforms
    weight: 100
---

# lookup
Using key from current row, get values from a reference source

## Parameters

| name | Type | Description |
| --- | --- | --- |
| replace | string (field path) | Field to replace | 
| lookup | string (template string) | Key to use for looking up data |
| copy | map[string]string | Given `lookup` of structure, copy values (key) to row (value) |  
| tsv | TSVTable  | TSV translation table file | 
| json | JSONTable | JSON data file | 
| table | LookupTable | Inline lookup table | 

## Example

```yaml
    - lookup:
        json:
          input: "{{config.doseResponseFile}}"
          key: experiment_id
        lookup: "{{row.experiment_id}}"
        copy:
          curve: curve
```
---
title: sqliteLoad
menu:
  main:
    parent: inputs
    weight: 100
---

# sqliteLoad

Extract data from an sqlite file

## Parameters

| Name | Type | Description |
|-------|---|--------|
| input | string | Path to the SQLite file |
| query | string | SQL select statement based input |

## Example

```yaml

inputs:
  sqlQuery:
    sqliteLoad:
      input: "{{config.sqlite}}"
      query: "select * from drug_mechanism as a LEFT JOIN MECHANISM_REFS as b on a.MEC_ID=b.MEC_ID LEFT JOIN TARGET_COMPONENTS as c on a.TID=c.TID LEFT JOIN COMPONENT_SEQUENCES as d on c.COMPONENT_ID=d.COMPONENT_ID LEFT JOIN MOLECULE_DICTIONARY as e on a.MOLREGNO=e.MOLREGNO"

```
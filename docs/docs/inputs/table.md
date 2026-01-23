---
title: table
menu:
  main:
    parent: inputs
    weight: 100
---

# table

Extract data from tabular file, includiong TSV and CSV files. 

## Parameters

| Name | Type | Description |
|-------|---|--------|
| path     |   string   | File to be transformed |
|	rowSkip   |   int       | Number of header rows to skip | 
|	columns   |   []string  | Manually set names of columns |
|	extraColumns | string   |  Columns beyond originally declared columns will be placed in this array |
|	sep       |   string   | Separator \\t for TSVs or , for CSVs |


## Example

```yaml

params:
  gafFile:
    default: ../../source/go/goa_human.gaf.gz
    type: File

inputs:
  gafLoad:
    tableLoad:
      path: "{{params.gafFile}}"
      columns:
        - db
        - id
        - symbol
        - qualifier
        - goID
        - reference
        - evidenceCode
        - from
        - aspect
        - name
        - synonym
        - objectType
        - taxon
        - date
        - assignedBy
        - extension
        - geneProduct

```
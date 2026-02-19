---
title: Inputs
menu:
  main:
    identifier: inputs
    weight: 4
---

Every playbook has a section of **input loaders** – components that read raw data (files, APIs, databases, etc.) and convert it into Python objects for downstream steps.  
An *input* can accept user‑supplied values passed by the **params** section.

## Common input types

* `table` – extracts data from tabular files (TSV/CSV)  
* `avro` – loads an Avro file (see `docs/docs/inputs/avro.md`)  
* `json`, `csv`, `sql`, etc.

## Example – `table`

The `table` loader is a good starting point because it demonstrates the typical parameter set required by most inputs.  See the full specification in `docs/docs/inputs/table.md`:

```yaml
params:
  gafFile:
    type: File
    default: ../../source/go/goa_human.gaf.gz

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

When you run the playbook you can override any of these parameters, e.g.:

```bash
sifter run gatplaybook.yaml --param gafFile=/tmp/mydata.tsv
```

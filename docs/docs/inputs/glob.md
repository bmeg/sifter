---
title: glob
menu:
  main:
    parent: inputs
    weight: 100
---

# glob
Scan files using `*` based glob statement and open all files
as input.

## Parameters

| Name | Description |
|-------|--------|
| storeFilename | Store value of filename in parameter each row |
| input | Path of avro object file to transform |
| xmlLoad | xmlLoad configutation |
| tableLoad | Run transform pipeline on a TSV or CSV |
| jsonLoad | Run a transform pipeline on a multi line json file |
| avroLoad | Load data from avro file |

## Example

```yaml
inputs:
  pubmedRead:
    glob:
      input: "{{config.baseline}}/*.xml.gz"
      xmlLoad: {}

```
---
title: xmlLoad
menu:
  main:
    parent: inputs
    weight: 100
---

# xmlLoad
Load an XML file

## Parameters

| name | Description |
| --- | --- |
| input | Path to input file |

## Example

```yaml
inputs:
  loader:
    xmlLoad:
      input: "{{config.xmlPath}}"
```
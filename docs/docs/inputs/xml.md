---
title: xml
menu:
  main:
    parent: inputs
    weight: 100
---

# xml
Load an XML file

## Parameters

| name | Description |
| --- | --- |
| path | Path to input file |

## Example

```yaml
inputs:
  loader:
    xmlLoad:
      path: "{{params.xmlPath}}"
```
---
title: jsonLoad
menu:
  main:
    parent: inputs
    weight: 100
---

# jsonLoad
Load data from a JSON file. Default behavior expects a single dictionary per line. Each line is a seperate entry. The `multiline` parameter reads all of the lines of the files and returns a single object.

## Parameters

| name | Description |
| --- | --- |
| input | Path of JSON file to transform |
|	multiline | Load file as a single multiline JSON object |


## Example

```yaml
inputs:
  caseData:
    jsonLoad:
      input: "{{config.casesJSON}}"
```
---
title: from
menu:
  main:
    parent: transforms
    weight: 100
---

# from

Start a pipeline from a named input or another pipeline.

## Parameters

| Name | Type | Description |
| --- | --- | --- |
| source | string | Name of the input or pipeline to start from |

## Example

```yaml
pipelines:
  profileProcess:
    - from: profileReader
```
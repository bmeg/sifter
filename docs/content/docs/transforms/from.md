---
title: from
menu:
  main:
    parent: transforms
    weight: 100
---

# from

## Parmeters

Name of data source

## Example

```yaml


inputs:
  profileReader:
    tableLoad:
      input: "{{config.profiles}}"

pipelines:
  profileProcess:
    - from: profileReader

```
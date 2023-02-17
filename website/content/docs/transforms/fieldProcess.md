---
title: fieldProcess
menu:
  main:
    parent: transforms
    weight: 100
---


# fieldProcess

Create stream of objects based on the contents of a field. If the selected field is an array
each of the items in the array will become an independent row.

## Parameters

| name | Type | Description |
| --- | --- | --- |
| field | string | Name of field to be processed |
| mapping | map[string]string | Project templated values into child element |
| itemField | string | If processing an array of non-dict elements, create a dict as `{itemField:element}` |


## example

```yaml
    - fieldProcess:
        field: portions
        mapping:
          sample: "{{row.sample_id}}"
          project_id: "{{row.project_id}}"
```
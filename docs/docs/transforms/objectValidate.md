---
title: objectValidate
menu:
  main:
    parent: transforms
    weight: 100
---

# objectValidate

Use JSON schema to validate row contents

# parameters

| name | Type | Description |
| --- | --- | --- |
| title | string | Title of object to use for validation |
| schema | string | Path to JSON schema definition |

# example

```
    - objectValidate:
        title: Aliquot
        schema: "{{config.schema}}"
```
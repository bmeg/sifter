---
title: clean
menu:
  main:
    parent: transforms
    weight: 100
---

type CleanStep struct {
	Fields      []string `json:"fields" jsonschema_description:"List of valid fields that will be left. All others will be removed"`
	RemoveEmpty bool     `json:"removeEmpty"`
	StoreExtra  string   `json:"storeExtra"`
}


# clearn

Remove fields that don't appear in the desingated list.

## Parameters

| name | Type | Description |
| --- | --- | --- |
| fields | [] string | Fields to keep | 
| removeEmpty | bool | Fields with empty values will also be removed |
| storeExtra | string | Field name to store removed fields |

## Example

```yaml
    - clean:
        fields:
          - id
          - synonyms
```
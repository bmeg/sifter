---
title: filter
menu:
  main:
    parent: transforms
    weight: 100
---

# filter

Filter rows in stream using a number of different methods

## Parameters

| name | Type | Description |
| --- | --- | --- |
| field | string (field path) | Field used to match rows | 
| value | string (template string) | Template string to match against |
| match | string | String to match against | 
| check | string | How to check value, 'exists' or 'hasValue' |
| method | string | Method name |
| python | string | Python code string |
| gpython | string | Python code string run using (https://github.com/go-python/gpython) |

## Example

Field based match
```yaml
    - filter:
        field: table
        match: source_statistics
```


Check based match
```yaml
    - filter:
        field: uniprot
        check: hasValue
```
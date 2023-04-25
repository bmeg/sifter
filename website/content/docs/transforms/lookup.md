---
title: lookup
menu:
  main:
    parent: transforms
    weight: 100
---

# lookup
Using key from current row, get values from a reference source

## Parameters

| name | Type | Description |
| --- | --- | --- |
| replace | string (field path) | Field to replace | 
| lookup | string (template string) | Key to use for looking up data |
| copy | map[string]string | Copy values from record that was found by lookup. The Key/Value record uses the Key as the destination field and copies the field from the retrieved records using the field named in Value |  
| tsv | TSVTable  | TSV translation table file | 
| json | JSONTable | JSON data file | 
| table | LookupTable | Inline lookup table | 
| pipeline | PipelineLookup | Use output of a pipeline as a lookup table |

## Example

### JSON file based lookup

The JSON file defined by `config.doseResponseFile` is opened and loaded into memory, using the `experiment_id` field as a primary key. 

```yaml
    - lookup:
        json:
          input: "{{config.doseResponseFile}}"
          key: experiment_id
        lookup: "{{row.experiment_id}}"
        copy:
          curve: curve
```


### Pipeline output lookup

Prepare a table in the pipelines `tableGen`. Then in `recordProcess` use that table, indexed by the field `primary_key` and lookup the value `{{row.table_id}}` to copy in the contents of the `other_data` field from the table and add it to the row as `my_data`.

```yaml

pipelines:

  tableGen:
    - from: dataFile
    #some set of transforms to prepair data
    #records look like { "primary_key" : "bob", "other_data": "red" }

  recordProcess:
    - from: recordFile
    - lookup:
        pipeline:
          from: tableGen
          key: primary_key
        lookup: "{{row.table_id}}"
        copy:
          my_data: other_data

```

#### Example data:
tableGen
```yaml
{ "primary_key" : "bob", "other_data": "red" }
{ "primary_key" : "alice", "other_data": "blue" }
```

recordProcess input
```yaml
{"id" : "record_1", "table_id":"alice" }
{"id" : "record_2", "table_id":"bob" }
```

recordProcess output
```yaml
{"id" : "record_1", "table_id":"alice", "my_data" : "blue" }
{"id" : "record_2", "table_id":"bob", "my_data" : "red" }
```


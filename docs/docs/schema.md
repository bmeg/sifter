---
title: Schema
---

# Sifter Playbook Schema

This document provides a comprehensive description of the Sifter Playbook format, its input methods (extractors), and its transformation steps.

## Playbook Structure

A Playbook is a YAML file that defines an ETL pipeline.

| Field | Type | Description |
| :--- | :--- | :--- |
| `class` | string | Should be `sifter`. |
| `name` | string | Unique name of the playbook. |
| `docs` | string | Documentation string for the playbook. |
| `outdir` | string | Default output directory for emitted files. |
| `params` | map | Configuration variables with optional defaults and types (`File`, `Dir`). |
| `inputs` | map | Named extractor definitions. |
| `outputs` | map | Named outputs definitions. |
| `pipelines` | map | Named transformation pipelines (arrays of steps). |

---

## Parameters (`params`)


Parameters allow playbooks to be parameterized. They are defined in the `params` section of the playbook YAML file.

### Params Syntax
```yaml
params:
  variableName:
    type: File # or Dir
    default: "path/to/default"
```

### Supported Types
- `File`: Represents a file path
- `Dir`: Represents a directory path

```yaml
params:
  inputDir:
    type: Dir
    default: "./data/input"
  outputDir:
    type: Dir
    default: "./data/output"
  schemaFile:
    type: File
    default: "./config/schema.json"
```


## Input Methods (Extractors)

Extractors produce a stream of messages from various sources.

### `table`
Loads data from a delimited file (TSV/CSV).
- `path`: Path to the file.
- `rowSkip`: Number of header rows to skip.
- `columns`: Optional list of column names.
- `extraColumns`: Field name to store any columns beyond the declared ones.
- `sep`: Separator (default `\t` for TSVs, `,` for CSVs).

### `json`
Loads data from a JSON file (standard or line-delimited).
- `path`: Path to the file.
- `multiline`: Load file as a single multiline JSON object.

### `avro`
Loads data from an Avro object file.
- `path`: Path to the file.

### `xml`
Loads and parses XML data.
- `path`: Path to the file.
- `level`: Depth level to start breaking XML into discrete messages.

### `sqlite`
Loads data from a SQLite database.
- `path`: Path to the database file.
- `query`: SQL SELECT statement.

### `transpose`
Loads a TSV and transposes it (making rows from columns).
- `input`: Path to the file.
- `rowSkip`: Rows to skip.
- `sep`: Separator.
- `useDB`: Use a temporary disk database for large transpositions.

### `plugin` (Extractor)
Runs an external command that produces JSON messages to stdout.
- `commandLine`: The command to execute.

### `embedded` (Extractor)
Load data from embedded structure.
- No parameters required.

### `glob` (Extractor)
Scan files using `*` based glob statement and open all files as input.
- `path`: Path of avro object file to transform.
- `storeFilename`: Store value of filename in parameter each row.
- `xml`: xmlLoad data.
- `table`: Run transform pipeline on a TSV or CSV.
- `json`: Run a transform pipeline on a multi line json file.
- `avro`: Load data from avro file.

---

## Transformation Steps

Transformation pipelines are arrays of steps. Each step can be one of the following:

### Core Processing
- `from`: Start a pipeline from a named input or another pipeline.
- `emit`: Write messages to a JSON file. Fields: `name`, `useName` (bool).
- `objectValidate`: Validate messages against a JSON schema. Fields: `title`, `schema` (directory), `uri`.
- `debug`: Print message contents to stdout. Fields: `label`, `format`.
- `plugin` (Transform): Pipe messages through an external script via stdin/stdout. Fields: `commandLine`.

### Mapping and Projection
- `project`: Map templates into new fields. Fields: `mapping` (key-template pairs), `rename` (simple rename).
- `map`: Apply a Python/GPython function to each record. Fields: `method` (function name), `python` (code string), `gpython` (path or code).
- `flatMap`: Similar to `map`, but flattens list responses into multiple messages.
- `fieldParse`: Parse a string field (e.g. `key1=val1;key2=val2`) into individual keys. Fields: `field`, `sep`.
- `fieldType`: Cast fields to specific types (`int`, `float`, `list`). Represented as a map of `fieldName: type`.

### Filtering and Cleaning
- `filter`: Drop messages based on criteria. Fields: `field`, `value`, `match`, `check` (`exists`/`hasValue`/`not`), or `python`/`gpython` code.
- `clean`: Remove fields. Fields: `fields` (list of kept fields), `removeEmpty` (bool), `storeExtra` (target field for extras).
- `dropNull`: Remove fields with `null` values from a message.
- `distinct`: Only emit messages with a unique value once. Field: `value` (template).

### Grouping and Lookups
- `reduce`: Merge messages sharing a key. Fields: `field` (key), `method`, `python`/`gpython`, `init` (initial data).
- `accumulate`: Group all messages sharing a key into a list. Fields: `field` (key), `dest` (target list field).
- `lookup`: Join data from external files (TSV/JSON). Fields: `tsv`, `json`, `replace`, `lookup`, `copy` (mapping of fields to copy).
- `intervalIntersect`: Match genomic intervals. Fields: `match` (CHR), `start`, `end`, `field` (dest), `json` (source file).

### Specialized
- `hash`: Generate a hash of a field. Fields: `field` (dest), `value` (template), `method` (`md5`, `sha1`, `sha256`).
- `uuid`: Generate a UUID. Fields: `field`, `value` (seed), `namespace`.
- `graphBuild`: Convert messages into graph vertices and edges using schema definitions. Fields: `schema`, `title`.
- `tableWrite`: Write specific fields to a delimited output file. Fields: `output`, `columns`, `sep`, `header`, `skipColumnHeader`.
- `split`: Split a single message into multiple based on a list field.

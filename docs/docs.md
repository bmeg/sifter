
---
title: Overview
menu:
  main:
    identifier: overview
    weight: 1
---

# Sifter pipelines

Sifter pipelines process steams of nested JSON messages. Sifter comes with a number of 
file extractors that operate as inputs to these pipelines. The pipeline engine 
connects togeather arrays of transform steps into directed acylic graph that is processed
in parallel.

Example Message:

```json
{
  "firstName" : "bob",
  "age" : "25"
  "friends" : [ "Max", "Alex"]
}
```

Once a stream of messages are produced, that can be run through a transform
pipeline. A transform pipeline is an array of transform steps, each transform
step can represent a different way to alter the data. The array of transforms link
togeather into a pipe that makes multiple alterations to messages as they are
passed along. There are a number of different transform steps types that can
be done in a transform pipeline these include:

 - Projection: creating new fields using a templating engine driven by existing values
 - Filtering: removing messages
 - Programmatic transformation: alter messages using an embedded python interpreter
 - Table based field translation
 - Outputing the message as a JSON Schema checked object


# Script structure

# Pipeline File

An sifter pipeline file is in YAML format and describes an entire processing pipelines. 
If is composed of the following sections: `config`, `inputs`, `pipelines`, `outputs`. In addition,
for tracking, the file will also include `name` and `class` entries. 

```yaml

class: sifter
name: <script name>
outdir: <where output files should go, relative to this file>

config:
  <config key>: <config value>
  <config key>: <config value> 
  # values that are referenced in pipeline parameters for 
  # files will be treated like file paths and be 
  # translated to full paths

inputs:
  <input name>:
    <input driver>:
      <driver config>

pipelines:
  <pipeline name>:
    # all pipelines must start with a from step
    - from: <name of input or pipeline> 
    - <transform name>:
       <transform parameters>

outputs:
  <output name>:
    <output driver>:
      <driver config>

```


## Header
Each sifter file starts with a set of field to let the software know this is a sifter script, and not some random YAML file. There is also a `name` field for the script. This name will be used for output file creation and logging. Finally, there is an `outdir` that defines the directory where all output files will be placed. All paths are relative to the script file, so the `outdir` set to `my-results` will create the directory `my-results` in the same directory as the script file, regardless of where the sifter command is invoked. 
```yaml
class : sifter
name: <name of script>
outdir: <where files should be stored>
```

# Config and templating
The `config` section is a set of defined keys that are used throughout the rest of the script. 

Example config:
```
config:
  sqlite:  ../../source/chembl/chembl_33/chembl_33_sqlite/chembl_33.db
  uniprot2ensembl: ../../tables/uniprot2ensembl.tsv
  schema: ../../schema/
```

Various fields in the script file will be be parsed using a [Mustache](https://mustache.github.io/) template engine. For example, to access the various values within the config block, the template `{{config.sqlite}}`.


# Inputs
The input block defines the various data extractors that will be used to open resources and create streams of JSON messages for processing. The possible input engines include:
 - AVRO
 - JSON
 - XML
 - SQL-dump
 - SQLite
 - TSV/CSV
 - GLOB

For any other file types, there is also a plugin option to allow the user to call their own code for opening files.

# Pipeline
The `pipelines` defined a set of named processing pipelines that can be used to transform data. Each pipeline starts with a `from` statement that defines where data comes from. It then defines a linear set of transforms that are chained togeather to do processing. Pipelines may used `emit` steps to output messages to disk. The possible data transform steps include:
- Accumulate
- Clean
- Distinct
- DropNull
- Field Parse
- Field Process
- Field Type
- Filter
- FlatMap
- GraphBuild
- Hash
- JSON Parse
- Lookup
- Value Mapping
- Object Validation
- Project
- Reduce
- Regex
- Split
- UUID Generation

Additionally, users are able to define their one transform step types using the `plugin` step.

# Example script
```yaml
class: sifter

name: go
outdir: ../../output/go/

config:
  oboFile: ../../source/go/go.obo
  schema: ../../schema

inputs:
  oboData:
    plugin:
      commandLine: ../../util/obo_reader.py {{config.oboFile}}

pipelines:
  transform:
    - from: oboData
    - project:
        mapping:
          submitter_id: "{{row.id[0]}}"
          case_id: "{{row.id[0]}}"
          id: "{{row.id[0]}}"
          go_id: "{{row.id[0]}}"
          project_id: "gene_onotology"
          namespace: "{{row.namespace[0]}}"
          name: "{{row.name[0]}}"
    - map: 
        method: fix
        gpython: | 
          def fix(row):
            row['definition'] = row['def'][0].strip('"')
            if 'xref' not in row:
              row['xref'] = []
            if 'synonym' not in row:
              row['synonym'] = []
            return row
    - objectValidate:
        title: GeneOntologyTerm
        schema: "{{config.schema}}"
    - emit:
        name: term
```
# Introduction
SIFTER is an Extract Transform Load (ETL) platform that is designed to take
a variety of standard input sources, create a message streams and run a
set of transforms to create JSON schema validated output classes.
SIFTER is based based on implementing a Playbook that describes top level
Extractions, that can include downloads, file manipulation and finally reading
the contents of the files. Every extractor is meant to produce a stream of
MESSAGES for transformation. A message is a simple nested dictionary data structure.

Example Message:

```
{
  "firstName" : "bob",
  "age" : "25"
  "friends" : [ "Max", "Alex"]
}
```

Once a stream of messages are produced, that can be run through a TRANSFORM
pipeline. A transform pipeline is an array of transform steps, each transform
step can represent a different way to alter the data. The array of transforms link
togeather into a pipe that makes multiple alterations to messages as they are
passed along. There are a number of different transform steps types that can
be done in a transform pipeline these include:

 - Projection
 - Filtering
 - Programmatic transformation
 - Table based field translation
 - Outputing the message as a JSON Schema checked object


***
# Example Playbook
Our first task will be to convert a ZIP code TSV into a set of county level
entries.

The input file looks like:

```
ZIP,COUNTYNAME,STATE,STCOUNTYFP,CLASSFP
36003,Autauga County,AL,01001,H1
36006,Autauga County,AL,01001,H1
36067,Autauga County,AL,01001,H1
36066,Autauga County,AL,01001,H1
36703,Autauga County,AL,01001,H1
36701,Autauga County,AL,01001,H1
36091,Autauga County,AL,01001,H1
```

First is the header of the Playbook. This declares the
unique name of the playbook and it's output directory.

```
name: zipcode_map
outdir: ./
docs: Converts zipcode TSV into graph elements
```

Next the configuration is declared. In this case the only input is the zipcode TSV.
There is a default value, so the playbook can be invoked without passing in
any parameters. However, to apply this playbook to a new input file, the
input parameter `zipcode` could be used to define the source file.

```
config:
  schema:
    type: Dir
    default: ../covid19_datadictionary/gdcdictionary/schemas/
  zipcode:
    type: File
    default: ../data/ZIP-COUNTY-FIPS_2017-06.csv
```

The `inputs` section declares data input sources. In this playbook, there is 
only one input, which is to run the table loader. 
```
inputs:
  tableLoad:
    input: "{{config.zipcode}}"
    sep: ","
```

Tableload operaters of the input file that was originally passed in using the
`inputs` stanza. SIFTER string parsing is based on mustache template system.
To access the string passed in the template is `{{config.zipcode}}`.
The seperator in the file input file is a `,` so that is also passed in as a
parameter to the extractor.


The `tableLoad` extractor opens up the TSV and generates a one message for
every row in the file. It uses the header of the file to map the column values
into a dictionary. The first row would produce the message:

```
{
    "ZIP" : "36003",
    "COUNTYNAME" : "Autauga County",
    "STATE" : "AL",
    "STCOUNTYFP" : "01001",
    "CLASSFP" : "H1"
}
```

The stream of messages are then passed into the steps listed in the `transform`
section of the tableLoad extractor.

For the current tranform, we want to produce a single entry per `STCOUNTYFP`,
however, the file has a line per `ZIP`. We need to run a `reduce` transform,
that collects rows togeather using a field key, which in this case is `"{{row.STCOUNTYFP}}"`,
and then runs a function `merge` that takes two messages, merges them togeather
and produces a single output message.

The two messages:

```
{ "ZIP" : "36003", "COUNTYNAME" : "Autauga County", "STATE" : "AL", "STCOUNTYFP" : "01001", "CLASSFP" : "H1"}
{ "ZIP" : "36006", "COUNTYNAME" : "Autauga County", "STATE" : "AL", "STCOUNTYFP" : "01001", "CLASSFP" : "H1"}
```

Would be merged into the message:

```
{ "ZIP" : ["36003", "36006"], "COUNTYNAME" : "Autauga County", "STATE" : "AL", "STCOUNTYFP" : "01001", "CLASSFP" : "H1"}
```

The `reduce` transform step uses a block of python code to describe the function.
The `method` field names the function, in this case `merge` that will be used
as the reduce function.

```
  zipReduce:
    - from: zipcode
    - reduce:
        field: STCOUNTYFP
        method: merge
        python: >
          def merge(x,y):
            a = x.get('zipcodes', []) + [x['ZIP']]
            b = y.get('zipcodes', []) + [y['ZIP']]
            x['zipcodes'] = a + b
            return x
```

The original messages produced by the loader have all of the information required
by the `summary_location` object type as described by the JSON schema that was linked
to in the header stanza. However, the data is all under the wrong field names.
To remap the data, we use a `project` tranformation that uses the template engine
to project data into new files in the message. The template engine has the current
message data in the value `row`. So the value
`FIPS:{{row.STCOUNTYFP}}` is mapped into the field `id`.

```
  - project:
      mapping:
        id: "FIPS:{{row.STCOUNTYFP}}"
        province_state: "{{row.STATE}}"
        summary_locations: "{{row.STCOUNTYFP}}"
        county: "{{row.COUNTYNAME}}"
        submitter_id: "{{row.STCOUNTYFP}}"
        type: summary_location
        projects: []
```

Using this projection, the message:

```
{
  "ZIP" : ["36003", "36006"],
  "COUNTYNAME" : "Autauga County",
  "STATE" : "AL",
  "STCOUNTYFP" : "01001",
  "CLASSFP" : "H1"
}
```

would become

```
{
  "id" : "FIPS:01001",
  "province_state" : "AL",
  "summary_locations" : "01001",
  "county" : "Autauga County",
  "submitter_id" : "01001",
  "type" : "summary_location"
  "projects" : [],
  "ZIP" : ["36003", "36006"],
  "COUNTYNAME" : "Autauga County",
  "STATE" : "AL",
  "STCOUNTYFP" : "01001",
  "CLASSFP" : "H1"
}
```

Now that the data has been remapped, we pass the data into the 'objectCreate'
transformation, which will read in the schema for `summary_location`, check the
message to make sure it matches and then output it.

```
  - objectCreate:
        class: summary_location
```


Outputs

To create an output table, with two columns connecting
`ZIP` values to `STCOUNTYFP` values. The `STCOUNTYFP` is a county level FIPS
code, used by the census office. A single FIPS code my contain many ZIP codes,
and we can use this table later for mapping ids when loading the data into a database.

```
outputs:
  zip2fips:
    tableWrite:
      from: 
      output: zip2fips
      columns:
        - ZIP
        - STCOUNTYFP
```


***
# File Format
A Playbook is a YAML file, that links a schema to a series of extractors that
in turn, can run several transforms to emit objects that are checked against
the schema.


***
## Playbook

The Playbook represents a single ETL pipeline that takes multiple inputs
and turns them into multiple output streams. It can take a set of inputs
then run a sequential set of extraction steps.


 -  class

> Type: *string* 

 -  name

> Type: *string* 

: Unique name of the playbook

 -  docs

> Type: *string* 

 -  outdir

> Type: *string* 

 -  config

> Type: *object*  of [ConfigVar](#configvar)


: Configuration for Playbook

 -  inputs

> Type: *object*  of [Extractor](#extractor)


: Steps of the transformation

 -  outputs

> Type: *object*  of [WriteConfig](#writeconfig)


 -  pipelines

> Type: *object* 


***
## ConfigVar

 -  name

> Type: *string* 

 -  type

> Type: *string* 

 -  default

> Type: *string* 


***
# Extraction Steps
Every playbook consists of a series of extraction steps. An extraction step
can be a data extractor that runs a transform pipeline.


***
## Extractor

This object represents a single extractor step. It has a field for each possible
extractor type, but only one is supposed to be filed in at a time.


 -  description

> Type: *string* 

: Human Readable description of step

 -  xmlLoad

 of [XMLLoadStep](#xmlloadstep)

 -  tableLoad

 of [TableLoadStep](#tableloadstep)

: Run transform pipeline on a TSV or CSV

 -  jsonLoad

 of [JSONLoadStep](#jsonloadstep)

: Run a transform pipeline on a multi line json file

 -  sqldumpLoad

 of [SQLDumpStep](#sqldumpstep)

: Parse the content of a SQL dump to find insert and run a transform pipeline

 -  gripperLoad

 of [GripperLoadStep](#gripperloadstep)

: Use a GRIPPER server to get data and run a transform pipeline

 -  avroLoad

 of [AvroLoadStep](#avroloadstep)

: Load data from avro file

 -  embedded

> Type: *array* 

 -  glob

 of [GlobLoadStep](#globloadstep)

 -  sqliteLoad

 of [SQLiteStep](#sqlitestep)

An array of Extractors, each defining a different extraction step

```
- desc: Untar the input file
  untar:
    input: "{{inputs.tar}}"
- desc: Loading Patient List
  tableLoad:
    input: data_clinical_patient.txt
    transform:
      ...
- desc: Loading Sample List
  tableLoad:
    input: data_clinical_sample.txt
    transform:
      ...
- fileGlob:
    files: [ data_RNA_Seq_expression_median.txt, data_RNA_Seq_V2_expression_median.txt ]
    steps:
      ...
```


***
## SQLDumpStep

 -  input

> Type: *string* 

: Path to the SQL dump file

 -  tables

> Type: *array* 

: Array of transforms for the different tables in the SQL dump


***
## TableLoadStep

 -  input

> Type: *string* 

: TSV to be transformed

 -  rowSkip

> Type: *integer* 

: Number of header rows to skip

 -  columns

> Type: *array* 

: Manually set names of columns

 -  extraColumns

> Type: *string* 

: Columns beyond originally declared columns will be placed in this array

 -  sep

> Type: *string* 

: Separator \t for TSVs or , for CSVs


***
## JSONLoadStep

 -  input

> Type: *string* 

: Path of multiline JSON file to transform

 -  transform

> Type: *array*  of [Step](#step)

: Transformation Pipeline

 -  multiline

> Type: *boolean* 

: Load file as a single multiline JSON object

```
- desc: Convert Census File
  jsonLoad:
    input: "{{inputs.census}}"
    transform:
      ...
```


***
## GripperLoadStep

Use a GRIPPER server to obtain data

 -  host

> Type: *string* 

: GRIPPER URL

 -  collection

> Type: *string* 

: GRIPPER collection to target


***
# Transform Pipelines
A tranform pipeline is a series of method to alter a message stream.


***
## ObjectCreateStep

Output a JSON schema described object

 -  class

> Type: *string* 

: Object class, should match declared class in JSON Schema

 -  schema

> Type: *string* 

: Directory with JSON schema files


***
## MapStep

Apply the sample function to every message

 -  method

> Type: *string* 

: Name of function to call

 -  python

> Type: *string* 

: Python code to be run

 -  gpython

> Type: *string* 

: Python code to be run using GPython

The `python` section defines the code, and the `method` parameter defines
which function from the code to call
```
- map:
    #fix weird formatting of zip code
    python: >
      def f(x):
        d = int(x['zipcode'])
        x['zipcode'] = "%05d" % (int(d))
        return x
    method: f
```


***
## ProjectStep

Project templates into fields in the message

 -  mapping

> Type: *object* 

: New fields to be generated from template

 -  rename

> Type: *object* 

: Rename field (no template engine)


```
- project:
    mapping:
      code: "{{row.project_id}}"
      programs: "{{row.program.name}}"
      submitter_id: "{{row.program.name}}"
      projects: "{{row.project_id}}"
      type: experiment
```


***
## LookupStep

Use a two column file to make values from one value to another.

 -  replace

> Type: *string* 

 -  tsv

 of [TSVTable](#tsvtable)

 -  json

 of [JSONTable](#jsontable)

 -  lookup

> Type: *string* 

 -  copy

> Type: *object* 

Starting with a table that maps state names to the two character state code:

```
North Dakota	ND
Ohio	OH
Oklahoma	OK
Oregon	OR
Pennsylvania	PA
```

The transform:

```
  - tableReplace:
      input: "{{inputs.stateTable}}"
      field: sub_region_1
```

Would change the message:

```
{ "sub_region_1" : "Oregon" }
```

to

```
{ "sub_region_1" : "OR" }
```


***
## RegexReplaceStep

Use a regular expression based replacement to alter a field

 -  field

> Type: *string* 

 -  regex

> Type: *string* 

 -  replace

> Type: *string* 

 -  dst

> Type: *string* 


```
- regexReplace:
    col: "{{row.attributes.Parent}}"
    regex: "^transcript:"
    replace: ""
    dst: transcript_id
```


***
## ReduceStep

 -  field

> Type: *string* 

 -  method

> Type: *string* 

 -  python

> Type: *string* 

 -  gpython

> Type: *string* 

 -  init

> Type: *object* 

```
  - reduce:
      field: "{{row.STCOUNTYFP}}"
      method: merge
      python: >
        def merge(x,y):
          a = x.get('zipcodes', []) + [x['ZIP']]
          b = y.get('zipcodes', []) + [y['ZIP']]
          x['zipcodes'] = a + b
          return x
```


***
## FilterStep

 -  field

> Type: *string* 

 -  value

> Type: *string* 

 -  match

> Type: *string* 

 -  check

> Type: *string* 

: How to check value, 'exists' or 'hasValue'

 -  method

> Type: *string* 

 -  python

> Type: *string* 

 -  gpython

> Type: *string* 

 -  steps

> Type: *array*  of [Step](#step)


Match based filtering:

```
  - filter:
      col: "{{row.tax_id}}"
      match: "9606"
      steps:
      - tableWrite:
```

Code based filters:

```
- filter:
    python: >
      def f(x):
        if 'FIPS' in x and len(x['FIPS']) > 0 and len(x['date']) > 0:
          return True
        return False
    method: f
    steps:
      - objectCreate:
          class: summary_report
```


***
## DebugStep

Print out messages

 -  label

> Type: *string* 

```
- debug: {}
```


***
## FieldProcessStep

Table an array field from a message, split it into a series of
messages and run on child transform pipeline. The `mapping` field
allows you to take data from the parent message and map it into the
child messages.


 -  field

> Type: *string* 

 -  mapping

> Type: *object* 

 -  itemField

> Type: *string* 

: If processing an array of non-dict elements, create a dict as {itemField:element}

```
- fieldProcess:
    col: portions
    mapping:
      samples: "{{row.id}}"
```


***
## FieldParseStep

Take a param style string and parse it into independent elements in the message

 -  field

> Type: *string* 

 -  sep

> Type: *string* 

 -  assign

> Type: *string* 


The messages

```
{ "attributes" : "ID=CDS:ENSP00000419345;Parent=transcript:ENST00000486405;protein_id=ENSP00000419345" }
```

After the transform:

```
  - fieldParse:
      col: attributes
      sep: ";"
```

Becomes:
```
{
  "attributes" : "ID=CDS:ENSP00000419345;Parent=transcript:ENST00000486405;protein_id=ENSP00000419345",
  "ID" : "CDS:ENSP00000419345",
  "Parent" : "transcript:ENST00000486405",
  "protein_id" : "ENSP00000419345"
}
```


***
## AccumulateStep

 -  field

> Type: *string* 

: Field to use for group definition

 -  dest

> Type: *string* 

## AvroLoadStep

 -  input

> Type: *string* 

: Path of avro object file to transform

## CleanStep

 -  fields

> Type: *array* 

: List of valid fields that will be left. All others will be removed

 -  removeEmpty

> Type: *boolean* 

 -  storeExtra

> Type: *string* 

## CommandLineTemplate

 -  template

> Type: *string* 

 -  outputs

> Type: *array* 

 -  inputs

> Type: *array* 

## DistinctStep

 -  value

> Type: *string* 

 -  steps

> Type: *array*  of [Step](#step)

## EdgeRule

 -  prefixFilter

> Type: *boolean* 

 -  blankFilter

> Type: *boolean* 

 -  toPrefix

> Type: *string* 

 -  sep

> Type: *string* 

 -  idTemplate

> Type: *string* 

## EmitStep

 -  name

> Type: *string* 

## GlobLoadStep

 -  storeFilename

> Type: *string* 

 -  input

> Type: *string* 

: Path of avro object file to transform

 -  xmlLoad

 of [XMLLoadStep](#xmlloadstep)

 -  tableLoad

 of [TableLoadStep](#tableloadstep)

: Run transform pipeline on a TSV or CSV

 -  jsonLoad

 of [JSONLoadStep](#jsonloadstep)

: Run a transform pipeline on a multi line json file

 -  avroLoad

 of [AvroLoadStep](#avroloadstep)

: Load data from avro file

## GraphBuildStep

 -  schema

> Type: *string* 

 -  class

> Type: *string* 

 -  idPrefix

> Type: *string* 

 -  idTemplate

> Type: *string* 

 -  idField

> Type: *string* 

 -  filePrefix

> Type: *string* 

 -  sep

> Type: *string* 

 -  fields

> Type: *object*  of [EdgeRule](#edgerule)


 -  flat

> Type: *boolean* 

## HashStep

 -  field

> Type: *string* 

 -  value

> Type: *string* 

 -  method

> Type: *string* 

## JSONTable

 -  input

> Type: *string* 

 -  value

> Type: *string* 

 -  key

> Type: *string* 

## SQLiteStep

 -  input

> Type: *string* 

: Path to the SQLite file

 -  query

> Type: *string* 

: SQL select statement based input

## SnakeFileWriter

 -  from

> Type: *string* 

 -  commands

> Type: *array*  of [CommandLineTemplate](#commandlinetemplate)

## Step

 -  from

> Type: *string* 

 -  fieldParse

 of [FieldParseStep](#fieldparsestep)

: fieldParse to run

 -  fieldType

> Type: *object* 

: Change type of a field (ie string -> integer)

 -  objectCreate

 of [ObjectCreateStep](#objectcreatestep)

: Create a JSON schema based object

 -  emit

 of [EmitStep](#emitstep)

: Write to unstructured JSON file

 -  filter

 of [FilterStep](#filterstep)

 -  clean

 of [CleanStep](#cleanstep)

 -  debug

 of [DebugStep](#debugstep)

: Print message contents to stdout

 -  regexReplace

 of [RegexReplaceStep](#regexreplacestep)

 -  project

 of [ProjectStep](#projectstep)

: Run a projection mapping message

 -  map

 of [MapStep](#mapstep)

: Apply a single function to all records

 -  reduce

 of [ReduceStep](#reducestep)

 -  distinct

 of [DistinctStep](#distinctstep)

 -  fieldProcess

 of [FieldProcessStep](#fieldprocessstep)

: Take an array field from a message and run in child transform

 -  lookup

 of [LookupStep](#lookupstep)

 -  hash

 of [HashStep](#hashstep)

 -  graphBuild

 of [GraphBuildStep](#graphbuildstep)

 -  accumulate

 of [AccumulateStep](#accumulatestep)

## TSVTable

 -  input

> Type: *string* 

 -  sep

> Type: *string* 

 -  value

> Type: *string* 

 -  key

> Type: *string* 

 -  header

> Type: *array* 

## TableWriter

 -  from

> Type: *string* 

 -  output

> Type: *string* 

: Name of file to create

 -  columns

> Type: *array* 

: Columns to be written into table file

 -  sep

> Type: *string* 

## WriteConfig

 -  tableWrite

 of [TableWriter](#tablewriter)

 -  snakefile

 of [SnakeFileWriter](#snakefilewriter)

## XMLLoadStep

 -  input

> Type: *string* 


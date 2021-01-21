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

First is the header of the Playbook. This declared that is a playbook, the
unique name of the playbook and the directory where the target schema files
can be found.

```
class: Playbook

name: zipcode_map
schema: ../covid19_datadictionary/gdcdictionary/schemas/

```

Next the inputs are declared. In this case the only input is the zipcode TSV.
There is a default value, so the playbook can be invoked without passing in
any parameters. However, to apply this playbook to a new input file, the
input parameter `zipcode` could be used to define the source file.

```

inputs:
  zipcode:
    type: File
    default: ../data/ZIP-COUNTY-FIPS_2017-06.csv
```

The `steps` section defines the sequence extraction steps that can be taken.
In this program, there is only one extraction step, which is to run the table
loader. The table loader then has a transform pipeline that is applied to the
data extracted from the file.

Tableload operaters of the input file that was originally passed in using the
`inputs` stanza. SIFTER string parsing is based on mustache template system.
To access the string passed in the template is `{{inputs.zipcode}}`.
The seperator in the file input file is a `,` so that is also passed in as a
parameter to the extractor.

```
steps:
  - desc: Convert ZIP code File
    tableLoad:
      input: "{{inputs.zipcode}}"
      sep: ","
      transform:
        .......
```

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

The first step is to create an output table, with two columns connecting
`ZIP` values to `STCOUNTYFP` values. The `STCOUNTYFP` is a county level FIPS
code, used by the census office. A single FIPS code my contain many ZIP codes,
and we can use this table later for mapping ids when loading the data into a database.

```
        - tableWrite:
            output: zip2fips
            columns:
              - ZIP
              - STCOUNTYFP
```

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


 -  name

> Type: *string* 

: Unique name of the playbook

 -  inputs

> Type: *object*  of [Input](#input)


: Optional inputs to Playbook

 -  outputs

> Type: *array*  of [Output](#output)

: Additional file created by Playbook

 -  schema

> Type: *string* 

: Name of directory with library of Gen3/JSON Schema files

 -  class

> Type: *string* 

: Notation for file inspection, set as 'Playbook'

 -  steps

> Type: *array*  of [Extractor](#extractor)

: Steps of the transformation


***
## Input

 -  type

> Type: *string* 

 -  default

> Type: *string* 

 -  source

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

 -  download

 of [DownloadStep](#downloadstep)

: Download a File

 -  untar

 of [UntarStep](#untarstep)

: Untar a file

 -  xmlLoad

 of [XMLLoadStep](#xmlloadstep)

 -  transposeFile

 of [TransposeFileStep](#transposefilestep)

: Take a matrix TSV and transpose it (row become columns)

 -  tableLoad

 of [TableLoadStep](#tableloadstep)

: Run transform pipeline on a TSV or CSV

 -  jsonLoad

 of [JSONLoadStep](#jsonloadstep)

: Run a transform pipeline on a multi line json file

 -  sqldumpLoad

 of [SQLDumpStep](#sqldumpstep)

: Parse the content of a SQL dump to find insert and run a transform pipeline

 -  fileGlob

 of [FileGlobStep](#fileglobstep)

: Scan a directory and run a ETL pipeline on each of the files

 -  script

 of [ScriptStep](#scriptstep)

: Execute a script

 -  digLoad

 of [DigLoadStep](#digloadstep)

: Use a GRIP Dig server to get data and run a transform pipeline

 -  avroLoad

 of [AvroLoadStep](#avroloadstep)

: Load data from avro file

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
## TransposeFileStep

 -  input

> Type: *string* 

: TSV to transpose

 -  output

> Type: *string* 

: Where transpose output should be stored

 -  lineSkip

> Type: *integer* 

: Number of header lines to skip

 -  lowMem

> Type: *boolean* 

: Do transpose without caching matrix in memory. Takes longer but works on large files


```
  - desc: Transpose RNA file
    transposeFile:
      input: "{{inputs.rnaFile}}"
      output: data_RNA_Seq_expression_median_transpose.txt
```


***
## DownloadStep

 -  source

> Type: *string* 

 -  dest

> Type: *string* 

 -  output

> Type: *string* 


***
## UntarStep

 -  input

> Type: *string* 

: Path to TAR file

 -  strip

> Type: *integer* 

: Number of base directories to strip with untaring


```
  - desc: Untar
    untar:
      input: "{{inputs.tar}}"
```


***
## SQLDumpStep

 -  input

> Type: *string* 

: Path to the SQL dump file

 -  tables

> Type: *array*  of [TableTransform](#tabletransform)

: Array of transforms for the different tables in the SQL dump

 -  skipIfMissing

> Type: *boolean* 

: Option to skip without fail if input file does not exist


***
## TableTransform

 -  name

> Type: *string* 

: Name of the SQL file to transform

 -  transform

> Type: *array*  of [Step](#step)

: The transform pipeline


***
## ScriptStep

 -  dockerImage

> Type: *string* 

: Docker image the contains script environment

 -  command

> Type: *array* 

: Command line, written as an array, to be run

 -  commandLine

> Type: *string* 

: Command line to be run

 -  stdout

> Type: *string* 

: File to capture stdout

 -  workdir

> Type: *string* 

```
  - desc: Scrape GDC Projects
    script:
      dockerImage: bmeg/sifter-gdc-scan
      command: [/opt/gdc-scan.py, projects]
  - desc: Scrape GDC Cases
    script:
      dockerImage: bmeg/sifter-gdc-scan
      command: [/opt/gdc-scan.py, cases]
```


***
## TableLoadStep

 -  input

> Type: *string* 

: TSV to be transformed

 -  rowSkip

> Type: *integer* 

: Number of header rows to skip

 -  skipIfMissing

> Type: *boolean* 

: Skip without error if file missing

 -  columns

> Type: *array* 

: Manually set names of columns

 -  transform

> Type: *array*  of [Step](#step)

: Transform pipelines

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

 -  skipIfMissing

> Type: *boolean* 

: Skip without error if file does note exist

```
- desc: Convert Census File
  jsonLoad:
    input: "{{inputs.census}}"
    transform:
      ...
```


***
## FileGlobStep

 -  files

> Type: *array* 

: Array of files (with wildcards) to scan for

 -  limit

> Type: *integer* 

 -  inputName

> Type: *string* 

: variable name the file will be stored in when calling the extraction steps

 -  steps

> Type: *array*  of [Extractor](#extractor)

: Extraction pipeline to run


***
## DigLoadStep

Use a GRIP DIG server to obtain data

 -  host

> Type: *string* 

: DIG URL

 -  collection

> Type: *string* 

: DIG collection to target

 -  transform

> Type: *array*  of [Step](#step)

: The transform pipeline to run


***
# Transform Pipelines
A tranform pipeline is a series of method to alter a message stream.


***
## ObjectCreateStep

Output a JSON schema described object

 -  class

> Type: *string* 

: Object class, should match declared class in JSON Schema

 -  name

> Type: *string* 

: domain name of stream, to separate it from other output streams of the same output type


***
## MapStep

Apply the sample function to every message

 -  method

> Type: *string* 

: Name of function to call

 -  python

> Type: *string* 

: Python code to be run

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
## TableWriteStep

 -  output

> Type: *string* 

: Name of file to create

 -  columns

> Type: *array* 

: Columns to be written into table file

 -  sep

> Type: *string* 


***
## TableReplaceStep

Use a two column file to make values from one value to another.

 -  input

> Type: *string* 

 -  field

> Type: *string* 

 -  target

> Type: *string* 

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

 -  col

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

 -  match

> Type: *string* 

 -  check

> Type: *string* 

: How to check value, 'exists' or 'hasValue'

 -  method

> Type: *string* 

 -  python

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
## ForkStep

 -  transform

> Type: *array* 

```
  - desc: Loading ProjectData
    jsonLoad:
      input: out.projects.json
      transform:
        - fork:
            transform:
              -
                - project:
                    mapping:
                      code: "{{row.project_id}}"
                      programs: "{{row.program.name}}"
                - objectCreate:
                    class: project
              -
                - project:
                    mapping:
                      code: "{{row.project_id}}"
                      programs: "{{row.program.name}}"
                      submitter_id: "{{row.program.name}}"
                      projects: "{{row.project_id}}"
                      type: experiment
                - objectCreate:
                    class: experiment
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

 -  steps

> Type: *array*  of [Step](#step)

 -  mapping

> Type: *object* 

```
- fieldProcess:
    col: portions
    mapping:
      samples: "{{row.id}}"
    steps:
      - fieldProcess:
          ...
```


***
## FieldMapStep

Take a param style string and parse it into independent elements in the message

 -  col

> Type: *string* 

 -  sep

> Type: *string* 


The messages

```
{ "attributes" : "ID=CDS:ENSP00000419345;Parent=transcript:ENST00000486405;protein_id=ENSP00000419345" }
```

After the transform:

```
  - fieldMap:
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
## CacheStep

The results of long running functions can be stored in a database and only
calculated as needed.


 -  transform

> Type: *array*  of [Step](#step)

The address lookup function provided by the Census bureau is takes time to
run, so we use the `cache` step to define a subsection of the pipeline
that should be cached in a database, and if the transform is run again,
the results would be pulled out of the database.

```
 - cache:
     transform:
       - map:
           method: addressLookup
           python: >

             import json

             from urllib import request, parse

             def addressLookup(x):
               try:
                 address = x['query']
                 baseUrl = "https://geocoding.geo.census.gov/geocoder/locations/onelineaddress?format=json&benchmark=9&address="
                 out = request.urlopen(baseUrl + parse.quote(address))
                 data = json.loads(out.read())
                 x['addressLookup'] = data
               except:
                 pass
               return x
```


***
## AlleleIDStep

 -  prefix

> Type: *string* 

 -  genome

> Type: *string* 

 -  chromosome

> Type: *string* 

 -  start

> Type: *string* 

 -  end

> Type: *string* 

 -  reference_bases

> Type: *string* 

 -  alternate_bases

> Type: *string* 

 -  dst

> Type: *string* 


***
## AvroLoadStep

 -  input

> Type: *string* 

: Path of avro object file to transform

 -  transform

> Type: *array*  of [Step](#step)

: Transformation Pipeline

 -  skipIfMissing

> Type: *boolean* 

: Skip without error if file does note exist

## CleanStep

 -  fields

> Type: *array* 

: List of valid fields that will be left. All others will be removed

 -  removeEmpty

> Type: *boolean* 

 -  storeExtra

> Type: *string* 

## EmitStep

 -  name

> Type: *string* 

## JSONFileLookupStep

 -  input

> Type: *string* 

 -  field

> Type: *string* 

 -  key

> Type: *string* 

 -  Project

> Type: *object* 

## Output

 -  type

> Type: *string* 

: File type: File, ObjectFile, VertexFile, EdgeFile

 -  path

> Type: *string* 

## Step

 -  fieldMap

 of [FieldMapStep](#fieldmapstep)

: fieldMap to run

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

 -  alleleID

 of [AlleleIDStep](#alleleidstep)

: Generate a standardized allele hash ID

 -  project

 of [ProjectStep](#projectstep)

: Run a projection mapping message

 -  map

 of [MapStep](#mapstep)

: Apply a single function to all records

 -  reduce

 of [ReduceStep](#reducestep)

 -  fieldProcess

 of [FieldProcessStep](#fieldprocessstep)

: Take an array field from a message and run in child transform

 -  tableWrite

 of [TableWriteStep](#tablewritestep)

: Write out a TSV

 -  tableReplace

 of [TableReplaceStep](#tablereplacestep)

: Load in TSV to map a fields values

 -  tableLookup

 of [TableLookupStep](#tablelookupstep)

 -  jsonLookup

 of [JSONFileLookupStep](#jsonfilelookupstep)

 -  fork

 of [ForkStep](#forkstep)

: Take message stream and split into multiple child transforms

 -  cache

 of [CacheStep](#cachestep)

: Sub a child transform pipeline, caching the results in a database

## TableLookupStep

 -  input

> Type: *string* 

 -  sep

> Type: *string* 

 -  field

> Type: *string* 

 -  key

> Type: *string* 

 -  header

> Type: *array* 

 -  Project

> Type: *object* 

## XMLLoadStep

 -  input

> Type: *string* 

 -  transform

> Type: *array*  of [Step](#step)

 -  skipIfMissing

> Type: *boolean* 


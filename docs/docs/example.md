---
title: Example
menu:
  main:
    identifier: example
    weight: 3
---


# Example Pipeline
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

First is the header of the pipeline. This declares the
unique name of the pipeline and it's output directory.

```
name: zipcode_map
outdir: ./
docs: Converts zipcode TSV into graph elements
```

Next the configuration is declared. In this case the only input is the zipcode TSV.
There is a default value, so the pipeline can be invoked without passing in
any parameters. However, to apply this pipeline to a new input file, the
input parameter `zipcode` could be used to define the source file.

```
config:
  schema: ../covid19_datadictionary/gdcdictionary/schemas/
  zipcode: ../data/ZIP-COUNTY-FIPS_2017-06.csv
```

The `inputs` section declares data input sources. In this pipeline, there is 
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

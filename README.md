
# Sifter

Sifter is a Extract Tranform Load (ETL) engine. It can be used to
Extract from a number of different data resources, including TSV files, SQLDump
files and external databases. It includes a pipeline description language to
define a set of Transform steps to create object messages that can be
validated using a JSON schema data.

Finally, SIFTER has a loader module that takes JSON message streams and load them
into a property graph using rules described by GEN3 based JSON schema files.


## ETL Process

1) Download external artifacts (files, database dumps)
2) Transform elements into JsonSchema compliant object streams. Each stream is a
single file of "\n" delimited. File name os <prefix>.<class id>.json.gz
3) Graph Transform
3.1) Reformatted to fix GIDs, lookup unfinished edge ids
3.2) takes that 'link' commands from the Gen3 formatted JsonSchema files
to generated 'Vertex.json.gz' and 'Edge.json.gz' files
3.3) Check for vertices that are used on edges but missing from vertex files


## Example Extract/Transform Playbook

More detailed descriptions can be found in out [Playbook manual](Playbook.md)

```
class: sifter
name: census_2010

config:
  census: ../data/census_2010_byzip.json
  date: "2010-01-01"
  schema: ../covid19_datadictionary/gdcdictionary/schemas/

inputs:
  censusData:
    jsonLoad:
      input: "{{config.census}}"

pipelines:
  transform:
    - from: censusData
    - map:
        #fix weird formatting of zip code
        gpython: >
          def f(x):
            d = int(x['zipcode'])
            x['zipcode'] = "%05d" % (int(d))
            return x
        method: f
    - project:
        mapping:
          submitter_id: "{{row.geo_id}}:{{inputs.date}}"
          type: census_report
          date: "{{config.date}}"
          summary_location: "{{row.zipcode}}"
    - objectValidate:
        title: census_report
        schema: "{{config.schema}}"
```


## Running Sifter


```
sifter run examples/genome.yaml
```


## Python Exec

Sifter will run Python code, however for this to function, the python environment
needs to have GRPC install. To install, run:
```
pip install grpcio-tools
```

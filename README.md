
# Sifter

Sifter is a Extract Tranform Load (ETL) engine. It can be used to
Extract from a number of different data resources, including TSV files, SQLDump
files and external databases. It includes a pipeline description language to
define a set of Transform steps to create object messages that can be
validated using a JSON schema data.

Finally, SIFTER has a loader module that takes JSON message streams and load them
into a property graph using rules described by JsonHyperSchema.

## Example Extract/Transform Playbook

```
class: sifter
name: census_2010

params:
  census: 
    type: File
    default: ../data/census_2010_byzip.json
  date: 
    type: string
    default: "2010-01-01"
  schema: 
    type: path
    default: ../covid19_datadictionary/gdcdictionary/schemas/

inputs:
  censusData:
    json:
      path: "{{params.census}}"

outputs:
  validated:
    json:
      path: census_data.ndjson

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
          submitter_id: "{{row.geo_id}}:{{params.date}}"
          type: census_report
          date: "{{params.date}}"
          summary_location: "{{row.zipcode}}"
    - objectValidate:
        title: census_report
        schema: "{{params.schema}}"
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

## Go Tests

Run go tests with
```
go clean -testcache
go test ./test/... -v
```
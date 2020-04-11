
# Sifter
## ETL server for GRIP

Sifter is a prototype web service that manages load requests into a GRIP instance.

This is a prototype that is still under development.


## ETL Process

1) Download external artifacts (files, database dumps)
2) Transform elements into JsonSchema compliant object streams. Each stream is a
single file of "\n" delimited. File name os <prefix>.<class id>.json.gz
3) Graph Transform
3.1) Reformatted to fix GIDs, lookup unfinished edge ids
3.2) takes that 'link' commands from the Gen3 formatted JsonSchema files
to generated 'Vertex.json.gz' and 'Edge.json.gz' files
3.3) Check for vertices that are used on edges but missing from vertex files

## Dev notes
Example server setup

### Build website
```
cd interface
npm i
npm run build
```

### Quick static file server
```bash
go get github.com/m3ng9i/ran
./bin/ran -r test-data/ -l
```

### Turn on Mongo Server
```bash
docker run -p 27017:27017 mongo
```

### GRIP server
Mongo Config File `grip.yml`
```yaml
Database: mongo

MongoDB:
  URL: localhost:27017
  UseAggregationPipeline: true
```

Start GRIP Server
```bash
./bin/grip server -c grip.yml
```

### Sifter Server
```bash
./bin/sifter server --playbooks src/github.com/bmeg/sifter/examples/ --web src/github.com/bmeg/sifter/interface/build
```

### Post import request

Manifest file example (saved in `test-data` directory served by ran):
```
gdc/Aliquot.Vertex.json.gz
gdc/AliquotFor.Edge.json.gz
gdc/Case.Vertex.json.gz
gdc/Compound.Vertex.json.gz
gdc/InProgram.Edge.json.gz
gdc/InProject.Edge.json.gz
gdc/Program.Vertex.json.gz
gdc/Project.Vertex.json.gz
gdc/Sample.Vertex.json.gz
gdc/SampleFor.Edge.json.gz
gdc/TreatedWith.Edge.json.gz
```


Input file
```json
{
  "url" : "http://localhost:8080/gdc.manifest",
  "baseURL" : "http://localhost:8080/"
}
```

Post import request (into graph `test`)
```bash
curl  -H "Content-Type: text/plain" --data-binary @input.json http://localhost:8090/api/playbook/GraphManifest/test
```

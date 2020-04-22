
# Sifter
## ETL server for GRIP

Sifter is a prototype ETL engine for build property graphs

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

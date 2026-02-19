# SIFTER Project Documentation to Source Code Mapping

## Inputs

| Documentation File | Source Code File |
|-------------------|------------------|
| docs/docs/inputs/avro.md | extractors/avro_load.go |
| docs/docs/inputs/embedded.md | extractors/embedded.go |
| docs/docs/inputs/glob.md | extractors/glob_load.go |
| docs/docs/inputs/json.md | extractors/json_load.go |
| docs/docs/inputs/plugin.md | extractors/plugin_load.go |
| docs/docs/inputs/sqldump.md | extractors/sqldump_step.go |
| docs/docs/inputs/sqlite.md | extractors/sqlite_load.go |
| docs/docs/inputs/table.md | extractors/tabular_load.go |
| docs/docs/inputs/xml.md | extractors/xml_step.go |

## Transforms

| Documentation File | Source Code File |
|-------------------|------------------|
| docs/docs/transforms/accumulate.md | transform/accumulate.go |
| docs/docs/transforms/clean.md | transform/clean.go |
| docs/docs/transforms/debug.md | transform/debug.go |
| docs/docs/transforms/distinct.md | transform/distinct.go |
| docs/docs/transforms/fieldParse.md | transform/field_parse.go |
| docs/docs/transforms/fieldProcess.md | transform/field_process.go |
| docs/docs/transforms/fieldType.md | transform/field_type.go |
| docs/docs/transforms/filter.md | transform/filter.go |
| docs/docs/transforms/flatmap.md | transform/flat_map.go |
| docs/docs/transforms/from.md | transform/from.go |
| docs/docs/transforms/hash.md | transform/hash.go |
| docs/docs/transforms/lookup.md | transform/lookup.go |
| docs/docs/transforms/map.md | transform/mapping.go |
| docs/docs/transforms/objectValidate.md | transform/object_validate.go |
| docs/docs/transforms/plugin.md | transform/plugin.go |
| docs/docs/transforms/project.md | transform/project.go |
| docs/docs/transforms/reduce.md | transform/reduce.go |
| docs/docs/transforms/regexReplace.md | transform/regex.go |
| docs/docs/transforms/split.md | transform/split.go |
| docs/docs/transforms/uuid.md | transform/uuid.go |

## Outputs

| Documentation File | Source Code File |
|-------------------|------------------|
| docs/docs/outputs/graphBuild.md | playbook/output_graph.go |
| docs/docs/outputs/json.md | playbook/output_json.go |
| docs/docs/outputs/tableWrite.md | playbook/output_table.go |
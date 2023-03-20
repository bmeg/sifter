package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

var GraphExtensionTag = "json_schema_graph"

type GraphSchema struct {
	Classes  map[string]*jsonschema.Schema
	compiler *jsonschema.Compiler
}

func yamlLoader(s string) (io.ReadCloser, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	f := u.Path
	if runtime.GOOS == "windows" {
		f = strings.TrimPrefix(f, "/")
		f = filepath.FromSlash(f)
	}
	if strings.HasSuffix(f, ".yaml") {
		source, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Error reading file: %s", f)
			return nil, err
		}
		d := map[string]any{}
		yaml.Unmarshal(source, &d)
		schemaText, err := json.Marshal(d)
		if err != nil {
			log.Printf("Error translating file: %s", f)
			return nil, err
		}
		return io.NopCloser(strings.NewReader(string(schemaText))), nil
	}
	return os.Open(f)
}

/*
func isEdge(s string) bool {
	if strings.Contains(s, "_definitions.yaml#/to_many") {
		return true
	} else if strings.Contains(s, "_definitions.yaml#/to_one") {
		return true
	}
	return false
}
*/

var graphExtMeta = jsonschema.MustCompileString("graphExtMeta.json", `{
	"properties" : {
		"targets": {
			"type" : "array",
			"items" : {
				"type" : "object",
				"properties" : {
					"type" : {
						"type": "object",
						"properties" : {
							"$ref" : {
								"type" : "string"
							}
						}
					},
					"backref" : {
						"type": "string"
					}
				}
			}
		}
	}
}`)

type graphExtCompiler struct{}

type Target struct {
	Schema  *jsonschema.Schema
	Backref string
}

type GraphExtension struct {
	Targets []Target
}

func (s GraphExtension) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	//fmt.Printf("graph schema validate\n")
	return nil
}

func (graphExtCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if e, ok := m["targets"]; ok {
		if ea, ok := e.([]any); ok {
			out := GraphExtension{Targets: []Target{}}
			for i := range ea {
				if emap, ok := ea[i].(map[string]any); ok {
					if tval, ok := emap["type"]; ok {
						if tmap, ok := tval.(map[string]any); ok {
							if ref, ok := tmap["$ref"]; ok {
								if refStr, ok := ref.(string); ok {
									backRef := ""
									if bval, ok := emap["backref"]; ok {
										if bstr, ok := bval.(string); ok {
											backRef = bstr
										}
									}
									sch, err := ctx.CompileRef(refStr, "./", false)
									if err == nil {
										out.Targets = append(out.Targets, Target{
											Schema:  sch,
											Backref: backRef,
										})
									} else {
										return nil, err
									}
								}
							}
						}
					}
				}
			}
			return out, nil
			//return GraphExtension{emap}, nil
		}
	}
	return nil, nil
}

type LoadOpt struct {
	LogError func(uri string, err error)
}

func isObjectSchema(sch *jsonschema.Schema) bool {
	if sch != nil {
		for _, i := range sch.Types {
			if i == "object" {
				return true
			}
		}
	}
	return false
}

func isArraySchema(sch *jsonschema.Schema) bool {
	if sch != nil {
		for _, i := range sch.Types {
			if i == "array" {
				return true
			}
		}
	}
	return false
}

func ObjectScan(sch *jsonschema.Schema) []*jsonschema.Schema {
	out := []*jsonschema.Schema{}

	isObject := isObjectSchema(sch)
	if isObject {
		out = append(out, sch)
	}

	if sch.Ref != nil {
		out = append(out, ObjectScan(sch.Ref)...)
	}

	for _, i := range sch.AnyOf {
		out = append(out, ObjectScan(i)...)
	}

	return out
}

func Load(path string, opt ...LoadOpt) (GraphSchema, error) {

	jsonschema.Loaders["file"] = yamlLoader

	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	compiler.RegisterExtension(GraphExtensionTag, graphExtMeta, graphExtCompiler{})

	info, err := os.Stat(path)
	if err != nil {
		return GraphSchema{}, err
	}
	out := GraphSchema{Classes: map[string]*jsonschema.Schema{}, compiler: compiler}
	if info.IsDir() {
		files, _ := filepath.Glob(filepath.Join(path, "*.yaml"))
		if len(files) == 0 {
			return GraphSchema{}, fmt.Errorf("no schema files found")
		}
		for _, f := range files {
			if sch, err := compiler.Compile(f); err == nil {
				if sch.Title != "" {
					out.Classes[sch.Title] = sch
				} else {
					//log.Printf("Title not found: %s %#v", f, sch)
				}
			} else {
				for _, i := range opt {
					if i.LogError != nil {
						i.LogError(f, err)
					}
				}
			}
		}
	} else {
		if sch, err := compiler.Compile(path); err == nil {
			for _, obj := range ObjectScan(sch) {
				if obj.Title != "" {
					out.Classes[obj.Title] = obj
				}
			}
		} else {
			for _, i := range opt {
				if i.LogError != nil {
					i.LogError(path, err)
				}
			}
		}
	}
	return out, nil
}

func (s GraphSchema) GetClass(classID string) *jsonschema.Schema {
	if class, ok := s.Classes[classID]; ok {
		return class
	}
	var err error
	var sch *jsonschema.Schema
	if sch, err = s.compiler.Compile(classID); err == nil {
		return sch
	}
	log.Printf("compile error: %s", err)
	return nil
}

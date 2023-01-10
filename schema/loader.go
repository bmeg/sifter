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
			log.Printf("Error reading file")
			return nil, err
		}
		d := map[string]any{}
		yaml.Unmarshal(source, &d)
		schemaText, err := json.Marshal(d)
		if err != nil {
			log.Printf("Error translating file")
			return nil, err
		}
		return io.NopCloser(strings.NewReader(string(schemaText))), nil
	}
	return os.Open(f)
}

func isEdge(s string) bool {
	if strings.Contains(s, "_definitions.yaml#/to_many") {
		return true
	} else if strings.Contains(s, "_definitions.yaml#/to_one") {
		return true
	}
	return false
}

var graphExtMeta = jsonschema.MustCompileString("graphExtMeta.json", `{
	"properties" : {
		"reference_backrefs": {
			"type" : "object",
			"properties" : {},
			"additionalProperties" : true
		}
	}
}`)

type graphExtCompiler struct{}

type GraphExtension struct {
	Backrefs map[string]any
}

func (s GraphExtension) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	//fmt.Printf("graph schema validate\n")
	return nil
}

func (graphExtCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {

	if e, ok := m["reference_backrefs"]; ok {
		if emap, ok := e.(map[string]any); ok {
			return GraphExtension{emap}, nil
		}
	}
	return nil, nil
}

type LoadOpt struct {
	LogError func(uri string, err error)
}

func ObjectScan(sch *jsonschema.Schema) []*jsonschema.Schema {
	out := []*jsonschema.Schema{}

	isObject := false
	for _, i := range sch.Types {
		if i == "object" {
			isObject = true
		}
	}
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
	if sch, err := s.compiler.Compile(classID); err == nil {
		return sch
	} else {
		log.Printf("compile error: %s", err)
	}
	return nil
}

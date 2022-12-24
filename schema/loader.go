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

type GraphSchema struct {
	Classes map[string]*jsonschema.Schema
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
			return nil, err
		}
		d := map[string]any{}
		yaml.Unmarshal(source, &d)
		schemaText, err := json.Marshal(d)
		if err != nil {
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

var referenceMeta = jsonschema.MustCompileString("referenceMeta.json", `{
	"properties" : {
		"reference_type_enum": {
			"type": "string"
		},
		"reference_backref": {
			"type" : "string"
		}
	}
}`)

type referenceCompiler struct{}

type referenceSchema struct {
	typeEnum []string
	backRef  string
}

func (s referenceSchema) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	return nil
}

func (referenceCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	eString := ""
	if e, ok := m["reference_type_enum"]; ok {
		n, _ := e.(string)
		eString = n
	}
	brString := ""
	if e, ok := m["reference_backref"]; ok {
		n, _ := e.(string)
		brString = n
	}
	if eString == "" && brString == "" {
		return nil, nil
	}
	// nothing to compile, return nil
	return referenceSchema{[]string{eString}, brString}, nil
}

func Load(path string) (GraphSchema, error) {

	jsonschema.Loaders["file"] = yamlLoader

	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	compiler.RegisterExtension("reference_type_enum", referenceMeta, referenceCompiler{})

	files, _ := filepath.Glob(filepath.Join(path, "*.yaml"))
	if len(files) == 0 {
		return GraphSchema{}, fmt.Errorf("No schema files found")
	}
	out := GraphSchema{Classes: map[string]*jsonschema.Schema{}}
	for _, f := range files {
		if sch, err := compiler.Compile(f); err == nil {
			out.Classes[sch.Id] = sch
		} else {
			log.Printf("Error loading: %s", err)
		}
	}
	return out, nil
}

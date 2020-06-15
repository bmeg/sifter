package manager

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/extractors"
	"github.com/ghodss/yaml"
)

type StepConfig interface{}

type Loader interface {
	Load() chan gripql.GraphElement
}

type Input struct {
	Type string `json:"type"`
	Default string `json:"default"`
}

type Inputs map[string]Input

type Playbook struct {
	Name  string `json:"name" jsonschema_description:"Unique name of the playbook"`
	Inputs Inputs `json:"inputs,omitempty" jsonschema_description:"Optional inputs to Playbook"`
	Schema string `json:"schema" jsonschema_description:"Name of directory with library of Gen3/JSON Schema files"`
	Class string `json:"class" jsonschema_description:"Notation for file inspection, set as 'Playbook'"`
	Steps []extractors.Extractor `json:"steps" jsonschema_description:"Steps of the transformation"`
	path string
}

// Parse parses a YAML doc into the given Config instance.
func parse(raw []byte, conf *Playbook) error {
	return yaml.Unmarshal(raw, conf)
}

// ParseFile parses a Sifter playbook file, which is formatted in YAML,
// and returns a Playbook struct.
func ParseFile(relpath string, conf *Playbook) error {
	if relpath == "" {
		return nil
	}

	// Try to get absolute path. If it fails, fall back to relative path.
	path, abserr := filepath.Abs(relpath)
	if abserr != nil {
		path = relpath
	}

	// Read file
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config at path %s: \n%v", path, err)
	}

	// Parse file
	err = parse(source, conf)
	if err != nil {
		return fmt.Errorf("failed to parse config at path %s: \n%v", path, err)
	}

	conf.path = path
	return nil
}

// ParseDataFile parses input file
func ParseDataFile(path string, data *map[string]interface{}) error {

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read data at path %s: \n%v", path, err)
	}
	return yaml.Unmarshal(raw, data)
}

// ParseDataString parses input string
func ParseDataString(raw string, data *map[string]interface{}) error {
	return yaml.Unmarshal([]byte(raw), data)
}

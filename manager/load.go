package manager

import (
	//"io"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/steps"
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
	Name  string `json:"name"`
	Inputs Inputs `json:"inputs"`
	Schema string `json:"schema"`
	Class string `json:"class"`
	Steps []steps.Step `json:"steps"`
}

// Parse parses a YAML doc into the given Config instance.
func Parse(raw []byte, conf *Playbook) error {
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
	err = Parse(source, conf)
	if err != nil {
		return fmt.Errorf("failed to parse config at path %s: \n%v", path, err)
	}

	conf.Schema = filepath.Join(filepath.Dir(path), conf.Schema)

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

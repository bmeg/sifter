package playbook

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/extractors"
	"github.com/bmeg/sifter/transform"
	"github.com/bmeg/sifter/writers"
	"sigs.k8s.io/yaml"
)

type Loader interface {
	Load() chan gripql.GraphElement
}

type Playbook struct {
	Name      string                          `json:"name" jsonschema_description:"Unique name of the playbook"`
	Docs      string                          `json:"docs"`
	Outdir    string                          `json:"outdir"`
	Config    config.Config                   `json:"config,omitempty" jsonschema_description:"Configuration for Playbook"`
	Inputs    map[string]extractors.Extractor `json:"inputs" jsonschema_description:"Steps of the transformation"`
	Outputs   map[string]writers.WriteConfig  `json:"outputs"`
	Pipelines map[string]transform.Pipe       `json:"pipelines"`
	path      string
}

// Parse parses a YAML doc into the given Config instance.
func parse(raw []byte, conf *Playbook) error {
	return yaml.UnmarshalStrict(raw, conf)
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

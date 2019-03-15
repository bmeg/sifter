package playbook

import (
	//"io"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/bmeg/grip/gripql"
	"github.com/ghodss/yaml"
)

type StepConfig interface{}

type Loader interface {
	Load() chan gripql.GraphElement
}

type PrepStep struct {
	Download string
	Command  string
	Chdir    string
	ArgsCopy string
}

type EdgeCreationStep struct {
	To    string `json:"to"`
	From  string `json:"from"`
	Label string `json:"label"`
}

type DestVertexCreateStep struct {
	Gid   string `json:"gid"`
	Label string `json:"label"`
}

type ColumnReplaceStep struct {
	Column  string `json:"col"`
	Pattern string `json:"pattern"`
	Replace string `json:"replace"`
}

type ImportStep struct {
	Desc         string            `json:"desc"`
	MatrixLoad   *MatrixLoadStep   `json:"matrixLoad"`
	ManifestLoad *ManifestLoadStep `json:"manifestLoad"`
}

type Playbook struct {
	Class string       `json:"class"`
	Prep  []PrepStep   `json:"prep"`
	Steps []ImportStep `json:steps`
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
	return nil
}

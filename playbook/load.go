
package playbook

import (
  "github.com/ghodss/yaml"
  "path/filepath"
  "io/ioutil"
  "fmt"
)

type StepConfig interface {}

type PrepStep struct {
  Download string
  Command string
  Chdir string
}

type EdgeCreationStep struct {
  To string `json:"to"`
  From string `json:"from"`
  Label string `json:"label"`
}

type DestVertexCreateStep struct {
  Gid string `json:"gid"`
  Label string `json:"label"`
}

type ColumnReplaceStep struct {
  Column string `json:"col"`
  Pattern string `json:"pattern"`
  Replace string `json:"replace"`
}

type MatrixLoadStep struct {
  RowLabel string `json:"rowLabel"`
  RowPrefix string `json:"rowPrefix"`
  RowSkip  int `json:"rowSkip"`
  Exclude []string `json:"exclude"`
  Transpose bool `json:"transpose"`
  IndexCol int `json:"transpose"`
  NoVertex bool `json:"noVertex"`
  Edge  []EdgeCreationStep `json:"edge"`
  DestVertex []DestVertexCreateStep `json:"destVertex"`
  ColumnReplace []ColumnReplaceStep `json:"columnReplace"`
  ColumnExclude []string `json:"columnExclude"`
}

type ImportStep struct {
  Input string `json:"input"`
  Desc  string `json:"desc"`
  MatrixLoad *MatrixLoadStep `json:"matrixLoad"`
}

type Playbook struct {
  Class string `json:"class"`
  Prep  []PrepStep `json:"prep"`
  Steps  []ImportStep `json:steps`
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

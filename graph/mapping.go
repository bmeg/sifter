package graph

import (
	"fmt"
	"io/ioutil"
  "github.com/ghodss/yaml"

  "github.com/bmeg/sifter/evaluate"
)

type GraphMapping struct {
	Domains map[string]DomainMap `json:"domains"`
}

type DomainMap map[string]ObjectMap

type ObjectMap struct {
	IdTemplate string `json:"idTemplate"`
	Label      string `json:"label"`
}

func LoadMapping(path string) (*GraphMapping, error) {
	o := GraphMapping{}
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read data at path %s: \n%v", path, err)
	}
	if err := yaml.Unmarshal(raw, &o); err != nil {
		return nil, fmt.Errorf("failed to load graph mapping %s : %s", path, err)
	}
	return &o, nil
}

func (o ObjectMap) MapObject(d map[string]interface{}) map[string]interface{} {
  if o.IdTemplate != "" {
    sid, err := evaluate.ExpressionString(o.IdTemplate, nil, d)
    if err == nil {
      d["id"] = sid
    }
  }
	return d
}

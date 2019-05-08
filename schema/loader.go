
package schema

import (
  "log"
  "fmt"
  "io/ioutil"
  "path/filepath"
  "encoding/json"
  "github.com/ghodss/yaml"
)

type Link struct {
  Name       string   `json:"name"`
  Backref    string   `json:"backref"`
  Label      string   `json:"label"`
  TargetType string   `json:"target_type"`
  Required   bool     `json:"required"`
}

type Value struct {
  StringVal  string
  IntVal     int64
  BoolVal    bool
}

func (v *Value) UnmarshalJSON(data []byte) error {
  if err := json.Unmarshal(data, &v.StringVal); err == nil {
    return nil
  } else if err := json.Unmarshal(data, &v.IntVal); err == nil {
    return nil
  } else if err := json.Unmarshal(data, &v.BoolVal); err == nil {
    return nil
  }
  return fmt.Errorf("Unknown type: %s", data)
}


type Property struct {
  Type        TypeClass `json:"type"`
  Ref         string    `json:"$ref"`
  SystemAlias string    `json:"systemAlias"`
  Description string    `json:"description"`
  Enum        []Value   `json:"enum"`
  Default     Value     `json:"default"`
  Format      string    `json:"format"`
}


type PropertyElement struct {
  Element     Property
  Value       string
}

func (w *PropertyElement) UnmarshalJSON(data []byte) error {
  s := ""
  e := Property{}
  if err := json.Unmarshal(data, &e); err == nil {
    w.Element = e
    return nil
  }
  if err := json.Unmarshal(data, &s); err == nil {
    w.Value = s
    return nil
  }
  return fmt.Errorf("Property not element or string: %s", data)
}

type Properties map[string]PropertyElement

type Schema struct {
  Id     string  `json:"id"`
  Title  string  `json:"title"`
  Type   string  `json:"type"`
  Required []string `json:"required"`
  UniqueKeys [][]string `json:"required"`
  Links      []Link     `json:"links"`
  Props      Properties `json:"properties"`
}

type TypeClass struct {
  Type    string
  Types   []string
}

func (w *TypeClass) UnmarshalJSON(data []byte) error {
  if err := json.Unmarshal(data, &w.Type); err == nil {
    return nil
  } else if err := json.Unmarshal(data, &w.Types); err == nil {
    return nil
  }
  return fmt.Errorf("Found unknown: %s", data)
}

type Schemas struct {
  Classes  map[string]Schema
}

func Load(path string) Schemas {
  files, _ := filepath.Glob(filepath.Join(path, "*.yaml"))
  out := Schemas{Classes:map[string]Schema{}}
  for _, f := range files {
    if s, err := LoadSchema(f); err == nil {
      out.Classes[s.Id] = s
    } else {
      log.Printf("Error loading: %s", err)
    }
  }
  return out
}


func LoadSchema(path string) (Schema, error) {
  	raw, err := ioutil.ReadFile(path)
  	if err != nil {
  		return Schema{}, fmt.Errorf("failed to read data at path %s: \n%v", path, err)
  	}
    s := Schema{}
    if err := yaml.Unmarshal(raw, &s); err != nil {
      return Schema{}, fmt.Errorf("failed to read data at path %s: \n%v", path, err)
    }
    return s, nil
}

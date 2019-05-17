
package schema

import (
  "log"
  "fmt"
  "strings"
  "io/ioutil"
  "path/filepath"
  "encoding/json"
  "github.com/ghodss/yaml"
)

type Link struct {
  Name         string   `json:"name"`
  Backref      string   `json:"backref"`
  Label        string   `json:"label"`
  TargetType   string   `json:"target_type"`
  Multiplicity string   `json:multiplicity`
  Required     bool     `json:"required"`
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

func (w PropertyElement) MarshalJSON() ([]byte, error) {
  if w.Value != "" {
    return json.Marshal(w.Value)
  }
  return json.Marshal(w.Element)
}


type Properties map[string]PropertyElement

type Schema struct {
  Id     string  `json:"id"`
  Title  string  `json:"title"`
  Type   string  `json:"type"`
  Required []string `json:"required"`
  UniqueKeys [][]string `json:"uniqueKeys"`
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

func RefPath(basePath string, ref string) (string, string) {
  vs := strings.Split(ref, "#")
  var pPath string
  if len(vs[0]) > 0 {
    dir := filepath.Dir(basePath)
    pPath = filepath.Join(dir, vs[0])
  } else {
    pPath = basePath
  }
  return pPath, vs[1]
}

func LoadRef(basePath string, ref string, cls interface{}) error {

  pPath, pElem := RefPath(basePath, ref)

  raw, err := ioutil.ReadFile(pPath)
  if err != nil {
    return fmt.Errorf("failed to read data at path %s: \n%v", pPath, err)
  }
  pProps := map[string]interface{}{}
  if err := yaml.Unmarshal(raw, &pProps); err != nil {
    return fmt.Errorf("failed to file reference at path %s: \n%v", pPath, err)
  }
  fName := pElem[1:len(pElem)]
  if fData, ok := pProps[fName]; ok {
    sData, _ := yaml.Marshal(fData)
    if err := yaml.Unmarshal(sData, cls); err != nil {
        return err
    }
  }
  return nil
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
    if ref, ok := s.Props["$ref"]; ok {
      np := map[string]PropertyElement{}
      if err := LoadRef(path, ref.Value, &np); err == nil {
        refFile, _ := RefPath(path, ref.Value)
        for k, v := range np {
          if v.Element.Ref != "" {
            err := LoadRef(refFile, v.Element.Ref, &v.Element)
            if err != nil {
              log.Printf("Error: %s", err)
            }
            v.Element.Ref = ""
          }
          s.Props[k] = v
        }
      } else {
        log.Printf("Error: %s", err)
        return Schema{}, err
      }
    }
    for k, v := range s.Props {
      if v.Element.Ref != "" {
        log.Printf("External Load: %s %s %s", path, k, v.Element.Ref)
        err := LoadRef(path, v.Element.Ref, &v.Element)
        if err != nil {
          log.Printf("Error: %s", err)
        }
        log.Printf("Element: %s", v.Element)
      }
    }
    return s, nil
}

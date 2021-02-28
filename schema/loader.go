package schema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
)

type Link struct {
	Name         string `json:"name"`
	Backref      string `json:"backref,omitempty"`
	Label        string `json:"label"`
	TargetType   string `json:"target_type"`
	Multiplicity string `json:"multiplicity"`
	Required     bool   `json:"required"`
	Subgroup     []Link `json:"subgroup"`
}

type Value struct {
	StringVal *string
	IntVal    *int64
	BoolVal   *bool
}

func (v *Value) UnmarshalJSON(data []byte) error {
	var stringVal string
	var intVal int64
	var boolVal bool

	if err := json.Unmarshal(data, &stringVal); err == nil {
		v.StringVal = &stringVal
		return nil
	} else if err := json.Unmarshal(data, &intVal); err == nil {
		v.IntVal = &intVal
		return nil
	} else if err := json.Unmarshal(data, &boolVal); err == nil {
		v.BoolVal = &boolVal
		return nil
	}
	return fmt.Errorf("Unknown type: %s", data)
}

func (v Value) MarshalJSON() ([]byte, error) {
	if v.StringVal != nil {
		return json.Marshal(v.StringVal)
	}
	if v.IntVal != nil {
		return json.Marshal(v.IntVal)
	}
	if v.BoolVal != nil {
		return json.Marshal(v.BoolVal)
	}
	return json.Marshal(nil)
}

type Property struct {
	Type        TypeClass `json:"type"`
	Ref         string    `json:"$ref,omitempty"`
	SystemAlias string    `json:"systemAlias,omitempty"`
	Description string    `json:"description"`
	Enum        []Value   `json:"enum,omitempty"`
	Default     Value     `json:"default,omitempty"`
	Format      string    `json:"format,omitempty"`
}

type PropertyElement struct {
	Element Property
	Value   string
	AnyOf   []Property `json:"anyOf"`
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
	if err := json.Unmarshal(data, &w); err == nil {
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

type Edge struct {
	Label string `json:"label"`
	To    string `json:"to"`
	From  string `json:"from"`
}

type Properties map[string]PropertyElement

type Schema struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Type       string     `json:"type"`
	Required   []string   `json:"required"`
	UniqueKeys [][]string `json:"uniqueKeys"`
	Links      []Link     `json:"links"`
	Edge       *Edge      `json:"edge"`
	Props      Properties `json:"properties"`
}

type TypeClass struct {
	Type  string
	Types []string
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
	Classes map[string]Schema
}

func (s Schemas) Get(cls string) (Schema, bool) {
	c, ok := s.Classes[cls]
	return c, ok
}

func (s Schemas) GetClasses() []string {
	out := []string{}
	for k := range s.Classes {
		out = append(out, k)
	}
	return out
}

func Load(path string) (Schemas, error) {
	files, _ := filepath.Glob(filepath.Join(path, "*.yaml"))
	if len(files) == 0 {
		return Schemas{}, fmt.Errorf("No schema files found")
	}
	out := Schemas{Classes: map[string]Schema{}}
	for _, f := range files {
		if s, err := LoadSchema(f); err == nil {
			out.Classes[s.ID] = s
		} else {
			log.Printf("Error loading: %s", err)
		}
	}
	return out, nil
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
	fName := pElem[1:]
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
	for k := range s.Props {
		if s.Props[k].Element.Ref != "" {
			//log.Printf("External Load: %s %s %s", path, k, s.Props[k].Element.Ref)
			elm := Property{}
			err := LoadRef(path, s.Props[k].Element.Ref, &elm)
			if err != nil {
				log.Printf("Error: %s", err)
			}
			//We're overwritting the current record with the contents pointed to with
			//$ref, but we're leaving some fields, if they've been defined...
			if s.Props[k].Element.SystemAlias != "" {
				elm.SystemAlias = s.Props[k].Element.SystemAlias
			}
			s.Props[k] = PropertyElement{Element: elm}
			//s.Props[k].Element.Ref = ""
		}
	}
	return s, nil
}

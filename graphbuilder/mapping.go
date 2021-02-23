package graphbuilder

import (
	//"io"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/schema"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/evaluate"
	"github.com/ghodss/yaml"
	//"google.golang.org/protobuf/types/known/structpb"
)

type Mapping struct {
	Schema  string             `json:"schema" jsonschema_description:"Name of directory with library of Gen3/JSON Schema files"`
	Rules   map[string]MapRule `json:"rules"`
	RuleMap []RuleMapping      `json:"ruleMap"`
}

type RuleMapping struct {
	Name string `json:"name"`
	Rule string `json:"rule"`
}

type EdgeRule struct {
	PrefixFilter bool    `json:"prefixFilter"`
	BlankFilter  bool    `json:"blankFilter"`
	ToPrefix     string  `json:"toPrefix"`
	FromPrefix   string  `json:"fromPrefix"`
	Sep          *string `json:"sep"`
	IDTemplate   string  `json:"idTemplate"`
}

type MapRule struct {
	Class      string               `json:"class"`
	IDPrefix   string               `json:"idPrefix"`
	IDTemplate string               `json:"idTemplate"`
	IDField    string               `json:"idField"`
	FilePrefix string               `json:"filePrefix"`
	Sep        *string              `json:"sep"`
	OutEdges   map[string]*EdgeRule `json:"outEdges"`
	InEdges    map[string]*EdgeRule `json:"inEdges"`
}

func LoadMapping(path string) (*Mapping, error) {
	o := Mapping{}
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read data at path %s: \n%v", path, err)
	}
	if err := yaml.Unmarshal(raw, &o); err != nil {
		return nil, fmt.Errorf("failed to load graph mapping %s : %s", path, err)
	}
	absPath, _ := filepath.Abs(path)
	dirPath := filepath.Dir(absPath)
	schemaPath := filepath.Join(dirPath, o.Schema)
	o.Schema = schemaPath
	return &o, nil
}

func (m *Mapping) GetVertexPrefixes() []string {
	out := []string{}
	for _, d := range m.Rules {
		if d.IDPrefix != "" {
			out = append(out, d.IDPrefix)
		}
	}
	return out
}

func (m *Mapping) GetEdgeEndPrefixes() [][]string {
	out := [][]string{}
	for _, d := range m.Rules {
		for _, e := range d.InEdges {
			if e.FromPrefix != "" {
				out = append(out, []string{d.IDPrefix, e.FromPrefix})
				out = append(out, []string{e.FromPrefix, d.IDPrefix})
			}
		}
		for _, e := range d.OutEdges {
			if e.ToPrefix != "" {
				out = append(out, []string{d.IDPrefix, e.ToPrefix})
				out = append(out, []string{e.ToPrefix, d.IDPrefix})
			}
		}
	}
	return out
}

func (m *Mapping) HasRule(path string) bool {
	base := filepath.Base(path)
	for _, r := range m.RuleMap {
		if ok, _ := filepath.Match(r.Name, base); ok {
			return true
		}
	}
	return false
}

func (m *Mapping) GetRule(path string) *MapRule {
	base := filepath.Base(path)
	for _, r := range m.RuleMap {
		if ok, _ := filepath.Match(r.Name, base); ok {
			if o, ok := m.Rules[r.Rule]; ok {
				return &o
			}
		}
	}
	return nil
}

func (m *Mapping) GetOutputFilePrefix(path string) string {
	r := m.GetRule(path)
	if r != nil {
		inputs := map[string]interface{}{
			"path":     path,
			"basename": filepath.Base(path),
		}
		val, err := evaluate.ExpressionString(r.FilePrefix, inputs, nil)
		if err == nil {
			return val
		}
	}
	return ""
}

func prefixAdjust(id string, prefix string, sep *string, filter bool) (string, error) {
	if prefix == "" {
		return id, nil
	}
	if !strings.HasPrefix(id, prefix) {
		if filter {
			return id, fmt.Errorf("Mismatch prefix")
		}
		s := ":"
		if sep != nil {
			s = *sep
		}
		return prefix + s + id, nil
	}
	return id, nil
}

func edgeToMap(e *gripql.Edge) map[string]interface{} {
	d := e.Data.AsMap()
	if d == nil {
		d = map[string]interface{}{}
	}
	d["_to"] = e.To
	d["_from"] = e.From
	d["_label"] = e.Label
	return d
}

func (m *Mapping) Process(path string, in chan map[string]interface{}, sch schema.Schemas, emitter loader.GraphEmitter) {
	rule := m.GetRule(path)

	if rule == nil {
		for range in {
		}
		return
	}

	for obj := range in {
		if rule.IDField != "" {
			if x, ok := obj[rule.IDField]; ok {
				obj["id"] = x
			}
		}
		if rule.IDPrefix != "" {
			if id, ok := obj["id"]; ok {
				if idStr, ok := id.(string); ok {
					if !strings.HasPrefix(idStr, rule.IDPrefix) {
						obj["id"] = rule.IDPrefix + ":" + idStr
					}
				}
			}
		}
		if o, err := sch.Generate(rule.Class, obj); err == nil {
			for _, j := range o {
				if j.Vertex != nil {
					err := emitter.EmitVertex(j.Vertex)
					if err != nil {
						log.Printf("Emit Error: %s", err)
					}
				} else if j.OutEdge != nil || j.InEdge != nil {
					var edge *gripql.Edge
					if j.OutEdge != nil {
						edge = j.OutEdge
						if er, ok := rule.OutEdges[edge.Label]; ok {
							var err error
							if er.BlankFilter && edge.To == "" {
								edge = nil
							} else if edge.To, err = prefixAdjust(edge.To, er.ToPrefix, er.Sep, er.PrefixFilter); err != nil {
								edge = nil
							}
							if edge != nil && er.IDTemplate != "" {
								val, err := evaluate.ExpressionString(er.IDTemplate, nil, edgeToMap(edge))
								if err == nil {
									edge.Gid = val
								}
							}
						}
					}
					if j.InEdge != nil {
						edge = j.InEdge
						if er, ok := rule.InEdges[edge.Label]; ok {
							var err error
							if er.BlankFilter && edge.From == "" {
								edge = nil
							} else if edge.From, err = prefixAdjust(edge.From, er.FromPrefix, er.Sep, er.PrefixFilter); err != nil {
								edge = nil
							}
							if edge != nil && er.IDTemplate != "" {
								val, err := evaluate.ExpressionString(er.IDTemplate, nil, edgeToMap(edge))
								if err == nil {
									edge.Gid = val
								}
							}

						}
					}
					if edge != nil {
						err := emitter.EmitEdge(edge)
						if err != nil {
							log.Printf("Emit Error: %s", err)
						}
					}
				}
			}
		}
	}
}

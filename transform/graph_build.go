package transform

import (
	"fmt"
	"log"
	"strings"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/schema"
	"github.com/bmeg/sifter/task"
)

type EdgeRule struct {
	PrefixFilter bool    `json:"prefixFilter"`
	BlankFilter  bool    `json:"blankFilter"`
	ToPrefix     string  `json:"toPrefix"`
	Sep          *string `json:"sep"`
	IDTemplate   string  `json:"idTemplate"`
}

type GraphBuildStep struct {
	Schema     string               `json:"schema"`
	Class      string               `json:"class"`
	IDPrefix   string               `json:"idPrefix"`
	IDTemplate string               `json:"idTemplate"`
	IDField    string               `json:"idField"`
	FilePrefix string               `json:"filePrefix"`
	Sep        *string              `json:"sep"`
	Fields     map[string]*EdgeRule `json:"fields"`
	Flat       bool                 `json:"flat"`
}

type graphBuildProcess struct {
	config GraphBuildStep
	task   task.RuntimeTask
	sch    schema.GraphSchema
	class  string
}

func (ts GraphBuildStep) Init(task task.RuntimeTask) (Processor, error) {

	className, err := evaluate.ExpressionString(ts.Class, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}
	path, err := evaluate.ExpressionString(ts.Schema, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	sc, err := schema.Load(path)
	if err != nil {
		return nil, err
	}
	//force the two emitters to be created. nil messages don't get emitted
	//but the output file will be created
	task.Emit("vertex", nil)
	task.Emit("edge", nil)

	return &graphBuildProcess{ts, task, sc, className}, nil
}

func (ts GraphBuildStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	if ts.Schema != "" {
		for _, s := range evaluate.ExpressionIDs(ts.Schema) {
			out = append(out, config.Variable{Type: config.Dir, Name: config.TrimPrefix(s)})
		}
	}
	return out
}

func (ts *graphBuildProcess) Close() {}

func (ts *graphBuildProcess) Process(i map[string]interface{}) []map[string]interface{} {

	out := []map[string]any{}

	if o, err := ts.sch.Generate(ts.class, i); err == nil {
		for _, j := range o {
			if j.Vertex != nil {
				j.Vertex.Gid, _ = prefixAdjust(j.Vertex.Gid, ts.config.IDPrefix, ts.config.Sep, false)
				if j.Vertex.Gid == "" {
					//default behavior if GID is not configured
					if x, ok := j.Vertex.Data.AsMap()["id"]; ok {
						if xStr, ok := x.(string); ok {
							j.Vertex.Gid = xStr
						}
					}
				}
				err := ts.task.Emit("vertex", ts.vertexToMap(j.Vertex))
				if err != nil {
					log.Printf("Emit Error: %s", err)
				}
			} else if j.OutEdge != nil || j.InEdge != nil {
				var edge *gripql.Edge
				if j.OutEdge != nil {
					edge = j.OutEdge
					edge.From, _ = prefixAdjust(edge.From, ts.config.IDPrefix, ts.config.Sep, false)
					if er, ok := ts.config.Fields[j.Field]; ok {
						var err error
						if er.BlankFilter && edge.To == "" {
							edge = nil
						} else if edge.To, err = prefixAdjust(edge.To, er.ToPrefix, er.Sep, er.PrefixFilter); err != nil {
							edge = nil
						}
						if edge != nil && er.IDTemplate != "" {
							val, err := evaluate.ExpressionString(er.IDTemplate, nil, ts.edgeToMap(edge))
							if err == nil {
								edge.Gid = val
							}
						}
					} else {
						//default rules
					}
				}
				if j.InEdge != nil {
					edge = j.InEdge
					edge.To, _ = prefixAdjust(edge.To, ts.config.IDPrefix, ts.config.Sep, false)
					if er, ok := ts.config.Fields[j.Field]; ok {
						var err error
						if er.BlankFilter && edge.From == "" {
							edge = nil
						} else if edge.From, err = prefixAdjust(edge.From, er.ToPrefix, er.Sep, er.PrefixFilter); err != nil {
							edge = nil
						}
						if edge != nil && er.IDTemplate != "" {
							val, err := evaluate.ExpressionString(er.IDTemplate, nil, ts.edgeToMap(edge))
							if err == nil {
								edge.Gid = val
							}
						}
					} else {
						//default rules
					}
				}
				if edge != nil {
					err := ts.task.Emit("edge", ts.edgeToMap(edge))
					if err != nil {
						log.Printf("Emit Error: %s", err)
					}
				}
			}
		}
	}

	return out

}

func prefixAdjust(id string, prefix string, sep *string, filter bool) (string, error) {
	if prefix == "" {
		return id, nil
	}
	if !strings.HasPrefix(id, prefix) {
		if filter {
			return id, fmt.Errorf("mismatch prefix")
		}
		s := ":"
		if sep != nil {
			s = *sep
		}
		return prefix + s + id, nil
	}
	return id, nil
}

func (ts *graphBuildProcess) edgeToMap(e *gripql.Edge) map[string]interface{} {
	d := e.Data.AsMap()
	if d == nil {
		d = map[string]interface{}{}
	}
	if ts.config.Flat {
		if e.Gid != "" {
			d["_id"] = e.Gid
		}
		d["_to"] = e.To
		d["_from"] = e.From
		d["_label"] = e.Label
		return d
	}

	out := map[string]any{}
	if e.Gid != "" {
		out["gid"] = e.Gid
	}
	out["to"] = e.To
	out["from"] = e.From
	out["label"] = e.Label
	out["data"] = d
	return out
}

func (ts *graphBuildProcess) vertexToMap(v *gripql.Vertex) map[string]interface{} {
	d := v.Data.AsMap()
	if d == nil {
		d = map[string]interface{}{}
	}
	if ts.config.Flat {
		if v.Gid != "" {
			d["_id"] = v.Gid
		}
		d["_label"] = v.Label
		return d
	}
	out := map[string]any{}
	if v.Gid != "" {
		out["gid"] = v.Gid
	}
	out["label"] = v.Label
	out["data"] = d
	return out
}

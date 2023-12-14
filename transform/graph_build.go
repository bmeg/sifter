package transform

import (
	"log"

	schema "github.com/bmeg/jsonschemagraph/util"
	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type EdgeFix struct {
	Method  string `json:"method"`
	GPython string `json:"gpython"`
}

type GraphBuildStep struct {
	Schema  string   `json:"schema"`
	Title   string   `json:"title"`
	Clean   bool     `json:"clean"`
	Flat    bool     `json:"flat"`
	EdgeFix *EdgeFix `json:"edgeFix"`
}

type graphBuildProcess struct {
	config GraphBuildStep
	task   task.RuntimeTask
	sch    schema.GraphSchema
	class  string

	edgeFix evaluate.Processor
}

func (ts GraphBuildStep) Init(task task.RuntimeTask) (Processor, error) {

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
	task.Emit("vertex", nil, false)
	task.Emit("edge", nil, false)

	var edgeFix evaluate.Processor
	if ts.EdgeFix != nil {
		if ts.EdgeFix.GPython != "" {
			log.Printf("Init Map: %s", ts.EdgeFix.GPython)
			e := evaluate.GetEngine("gpython", task.WorkDir())
			c, err := e.Compile(ts.EdgeFix.GPython, ts.EdgeFix.Method)
			if err != nil {
				log.Printf("Compile Error: %s", err)
			}
			edgeFix = c
		}
	}
	return &graphBuildProcess{ts, task, sc, ts.Title, edgeFix}, nil
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

func (ts *graphBuildProcess) PoolReady() bool {
	return true
}

func (ts *graphBuildProcess) Close() {}

func (ts *graphBuildProcess) Process(i map[string]interface{}) []map[string]interface{} {

	out := []map[string]any{}

	if o, err := ts.sch.Generate(ts.class, i, ts.config.Clean); err == nil {
		for _, j := range o {
			if j.Vertex != nil {
				err := ts.task.Emit("vertex", ts.vertexToMap(j.Vertex), false)
				if err != nil {
					log.Printf("Emit Error: %s", err)
				}
			} else if j.OutEdge != nil || j.InEdge != nil {
				var edge *schema.Edge
				if j.OutEdge != nil {
					edge = j.OutEdge
				}
				if j.InEdge != nil {
					edge = j.InEdge
				}
				if edge != nil {
					edgeData := ts.edgeToMap(edge)
					if ts.edgeFix != nil {
						o, err := ts.edgeFix.Evaluate(edgeData)
						if err == nil {
							edgeData = o
						}
					}
					err := ts.task.Emit("edge", edgeData, false)
					if err != nil {
						log.Printf("Emit Error: %s", err)
					}
				}
			}
		}
	} else {
		log.Printf("Graphbuild %s error : %s", ts.config.Title, err)
	}

	return out

}

func (ts *graphBuildProcess) edgeToMap(e *schema.Edge) map[string]interface{} {
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

func (ts *graphBuildProcess) vertexToMap(v *schema.Vertex) map[string]interface{} {
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

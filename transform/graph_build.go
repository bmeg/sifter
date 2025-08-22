package transform

import (
	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/jsonschemagraph/graph"
	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type EdgeFix struct {
	Method  string     `json:"method"`
	GPython *CodeBlock `json:"gpython"`
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
	sch    graph.GraphSchema
	class  string

	edgeFix     evaluate.Processor
	objectCount int
	vertexCount int
	edgeCount   int
}

func (ts GraphBuildStep) Init(task task.RuntimeTask) (Processor, error) {

	path, err := evaluate.ExpressionString(ts.Schema, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	sc, err := graph.Load(path)
	if err != nil {
		return nil, err
	}
	//force the two emitters to be created. nil messages don't get emitted
	//but the output file will be created
	task.Emit("vertex", nil, false)
	task.Emit("edge", nil, false)

	var edgeFix evaluate.Processor
	if ts.EdgeFix != nil {
		if ts.EdgeFix.GPython != nil {
			ts.EdgeFix.GPython.SetBaseDir(task.BaseDir())
			logger.Debug("Init Map: %s", ts.EdgeFix.GPython)
			e := evaluate.GetEngine("gpython", task.WorkDir())
			c, err := e.Compile(ts.EdgeFix.GPython.String(), ts.EdgeFix.Method)
			if err != nil {
				logger.Error("Compile Error: %s", err)
			}
			edgeFix = c
		}
	}
	return &graphBuildProcess{ts, task, sc, ts.Title, edgeFix, 0, 0, 0}, nil
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

func (ts *graphBuildProcess) Close() {
	logger.Info("Graph Emit",
		"objects", ts.objectCount,
		"edges", ts.edgeCount,
		"vertices", ts.vertexCount,
		"class", ts.class)
}

func (ts *graphBuildProcess) Process(i map[string]interface{}) []map[string]interface{} {

	out := []map[string]any{}
	if o, err := ts.sch.Generate(ts.class, i, ts.config.Clean, map[string]any{}); err == nil {
		ts.objectCount++
		for i := range o {
			if o[i].Vertex != nil {
				ts.vertexCount++
				err := ts.task.Emit("vertex", ts.vertexToMap(o[i].Vertex), false)
				if err != nil {
					logger.Error("Emit Error: %s", err)
				}
			} else if o[i].Edge != nil {
				var edge *gripql.Edge = o[i].Edge
				if edge != nil {
					edgeData := ts.edgeToMap(edge)
					if ts.edgeFix != nil {
						o, err := ts.edgeFix.Evaluate(edgeData)
						if err == nil {
							edgeData = o
						}
					}
					ts.edgeCount++
					err := ts.task.Emit("edge", edgeData, false)
					if err != nil {
						logger.Error("Emit Error: %s", err)
					}
				}
			}
		}
	} else {
		logger.Error("Graphbuild %s error : %s", ts.config.Title, err)
	}

	return out

}

func (ts *graphBuildProcess) edgeToMap(e *gripql.Edge) map[string]interface{} {
	d := e.Data.AsMap()
	if d == nil {
		d = map[string]interface{}{}
	}
	if e.Id != "" {
		d["_id"] = e.Id
	}
	d["_to"] = e.To
	d["_from"] = e.From
	d["_label"] = e.Label
	return d
}

func (ts *graphBuildProcess) vertexToMap(v *gripql.Vertex) map[string]interface{} {
	d := v.Data.AsMap()
	if d == nil {
		d = map[string]interface{}{}
	}
	if v.Id != "" {
		d["_id"] = v.Id
	}
	d["_label"] = v.Label
	return d
}

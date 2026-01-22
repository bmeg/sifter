package playbook

import (
	"path/filepath"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/jsonschemagraph/graph"
	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/playbook/refs"
	"github.com/bmeg/sifter/task"
)

type EdgeFix struct {
	Method  string          `json:"method"`
	GPython *refs.CodeBlock `json:"gpython"`
}

type OutputGraph struct {
	From    string   `json:"from"`
	Output  string   `json:"output"`
	Schema  string   `json:"schema"`
	Title   string   `json:"title"`
	Clean   bool     `json:"clean"`
	Flat    bool     `json:"flat"`
	EdgeFix *EdgeFix `json:"edgeFix"`
}

func (oj *OutputGraph) GetOutputs(task task.RuntimeTask) []string {
	output, err := evaluate.ExpressionString(oj.Output, task.GetConfig(), nil)
	if err != nil {
		return []string{}
	}
	outputPath := filepath.Join(task.OutDir(), output)
	logger.Debug("table output %s %s", task.OutDir(), output)
	return []string{outputPath + ".edge", outputPath + ".vertex"}
}

type graphBuildProcess struct {
	config OutputGraph
	task   task.RuntimeTask
	sch    graph.GraphSchema
	class  string

	edgeName    string
	verrtexName string

	edgeFix     evaluate.Processor
	objectCount int
	vertexCount int
	edgeCount   int
}

func (ts OutputGraph) Init(task task.RuntimeTask) (OutputProcessor, error) {

	path, err := evaluate.ExpressionString(ts.Schema, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	sc, err := graph.Load(path)
	if err != nil {
		return nil, err
	}

	output, err := evaluate.ExpressionString(ts.Output, task.GetConfig(), nil)

	//TODO: make this more flexible
	edgeName := output + ".edge.json.gz"
	vertexName := output + ".vertex.json.gz"

	//force the two emitters to be created. nil messages don't get emitted
	//but the output file will be created
	task.Emit(vertexName, nil)
	task.Emit(edgeName, nil)

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
	return &graphBuildProcess{
		config:      ts,
		task:        task,
		sch:         sc,
		edgeName:    edgeName,
		verrtexName: vertexName,
		class:       ts.Title,
		edgeFix:     edgeFix,
		objectCount: 0,
		vertexCount: 0,
		edgeCount:   0}, nil
}

func (ts OutputGraph) GetRequiredParams() []config.ParamRequest {
	out := []config.ParamRequest{}
	if ts.Schema != "" {
		for _, s := range evaluate.ExpressionIDs(ts.Schema) {
			out = append(out, config.ParamRequest{Type: "File", Name: config.TrimPrefix(s)})
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

func (ts *graphBuildProcess) Process(i map[string]interface{}) {

	if o, err := ts.sch.Generate(ts.class, i, ts.config.Clean, map[string]any{}); err == nil {
		ts.objectCount++
		for i := range o {
			if o[i].Vertex != nil {
				ts.vertexCount++
				err := ts.task.Emit(ts.verrtexName, ts.vertexToMap(o[i].Vertex))
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
					err := ts.task.Emit(ts.edgeName, edgeData)
					if err != nil {
						logger.Error("Emit Error: %s", err)
					}
				}
			}
		}
	} else {
		logger.Error("Graphbuild %s error : %s", ts.config.Title, err)
	}
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

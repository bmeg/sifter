package graphedit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/manager"

	"google.golang.org/protobuf/types/known/structpb"
)

func NewTask(workdir string, vertexEmit bool, graphOut loader.GraphEmitter) manager.RuntimeTask {
	return &FixTask{name: "graph-fix", workdir: workdir, out: graphOut, vertexEmit: vertexEmit}
}

type FixTask struct {
	name       string
	workdir    string
	out        loader.GraphEmitter
	vertexEmit bool
}

func (m *FixTask) AbsPath(p string) (string, error) {
	if !strings.HasPrefix(p, "/") {
		p = filepath.Join(m.workdir, p)
	}
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return a, nil
}

func (m *FixTask) Child(name string) manager.RuntimeTask {
	return &FixTask{workdir: m.workdir, name: m.name + ":" + name, out: m.out, vertexEmit: m.vertexEmit}
}

func (m *FixTask) GetInputs() map[string]interface{} {
	return map[string]interface{}{}
}

func (m *FixTask) TempDir() string {
	return "./"
}

func (m *FixTask) WorkDir() string {
	return "./"
}

func (m *FixTask) GetName() string {
	return m.name
}

func (m *FixTask) GetDataStore() (datastore.DataStore, error) {
	return nil, fmt.Errorf("Datastores not enabled for graph-fix")
}

func (m *FixTask) Emit(name string, out map[string]interface{}) error {
	if m.vertexEmit {
		v := gripql.Vertex{}
		if id, ok := out["gid"]; ok {
			if idStr, ok := id.(string); ok {
				v.Gid = idStr
			}
		}
		if label, ok := out["label"]; ok {
			if labelStr, ok := label.(string); ok {
				v.Label = labelStr
			}
		}
		if data, ok := out["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				v.Data, _ = structpb.NewStruct(dataMap)
			}
		}
		m.out.EmitVertex(&v)
	} else {
		e := gripql.Edge{}
		if id, ok := out["gid"]; ok {
			if idStr, ok := id.(string); ok {
				e.Gid = idStr
			}
		}
		if label, ok := out["label"]; ok {
			if labelStr, ok := label.(string); ok {
				e.Label = labelStr
			}
		}
		if to, ok := out["to"]; ok {
			if toStr, ok := to.(string); ok {
				e.To = toStr
			}
		}
		if from, ok := out["from"]; ok {
			if fromStr, ok := from.(string); ok {
				e.From = fromStr
			}
		}
		if data, ok := out["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				e.Data, _ = structpb.NewStruct(dataMap)
			}
		}

		m.out.EmitEdge(&e)
	}
	return nil
}

func (m *FixTask) EmitObject(prefix string, objClass string, e map[string]interface{}) error {
	return fmt.Errorf("Object emit not enabled in graph-fix")
}

func (m *FixTask) EmitTable(prefix string, columns []string, sep rune) loader.TableEmitter {
	return nil
}

func (m *FixTask) Close() {}

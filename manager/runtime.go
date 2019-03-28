package manager

import (
	"sync/atomic"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/emitter"
)

type Runtime struct {
	man    *Manager
	output emitter.Emitter
	dir    string
}

func (run *Runtime) NewTask(inputs map[string]interface{}) *Task {
	return &Task{run.man, run, run.dir, inputs}
}

func (run *Runtime) Close() {
	if run.output != nil {
		run.output.Close()
	}
}

func (run *Runtime) EmitVertex(v *gripql.Vertex) error {
	atomic.AddInt64(&run.man.VertexCount, 1)
	return run.output.EmitVertex(v)
}

func (run *Runtime) EmitEdge(e *gripql.Edge) error {
	atomic.AddInt64(&run.man.EdgeCount, 1)
	return run.output.EmitEdge(e)
}

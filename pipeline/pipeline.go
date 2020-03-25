package pipeline

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/emitter"
	"github.com/bmeg/sifter/schema"
)

type Runtime struct {
	//man         *Manager
	output      emitter.Emitter
	dir         string
	name        string
	Status      string
	VertexCount int64
	EdgeCount   int64
	StepCount   int64
	StepTotal   int64
	OutputCallback func(string, string) error
	Schemas     *schema.Schemas
}

func NewRuntime(output emitter.Emitter, dir string, name string) *Runtime {
	return &Runtime{output:output, dir:dir, name:name, Status:"Starting"}
}

func (run *Runtime) NewTask(inputs map[string]interface{}) *Task {
	return &Task{Runtime:run, Workdir:run.dir, Inputs:inputs}
}

func (run *Runtime) Close() {
	if run.output != nil {
		run.output.Close()
	}
	//run.man.DropRuntime(run.name)
}


func (run *Runtime) LoadSchema(path string) {
	a := schema.Load(path)
	run.Schemas = &a
}

func (run *Runtime) EmitVertex(v *gripql.Vertex) error {
	atomic.AddInt64(&run.VertexCount, 1)
	return run.output.EmitVertex(v)
}

func (run *Runtime) EmitEdge(e *gripql.Edge) error {
	atomic.AddInt64(&run.EdgeCount, 1)
	return run.output.EmitEdge(e)
}


func (run *Runtime) EmitObject(c string, o map[string]interface{}) error {
	return run.output.EmitObject(c,o)
}

func (m *Runtime) Printf(s string, x ...interface{}) {
	c := fmt.Sprintf(s, x...)
	log.Printf(c)
	m.Status = c
}

func (m *Runtime) GetCurrent() string {
	return m.Status
}

func (m *Runtime) GetVertexCount() int64 {
	return m.VertexCount
}

func (m *Runtime) GetEdgeCount() int64 {
	return m.EdgeCount
}

func (m *Runtime) GetStepNum() int64 {
	return m.StepCount
}

func (m *Runtime) GetStepTotal() int64 {
	return m.StepTotal
}

func (m *Runtime) SetStepCountTotal(i int64) {
	m.StepTotal = i
}

func (m *Runtime) AddStepCount(i int64) {
	atomic.AddInt64(&m.StepCount, i)
}

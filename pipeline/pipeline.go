package pipeline

import (
	"fmt"
	"log"
	"sync/atomic"

	//"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/emitter"
)

type Runtime struct {
	//man         *Manager
	output         emitter.Emitter
	dir            string
	name           string
	Status         string
	StepCount      int64
	StepTotal      int64
	OutputCallback func(string, string) error
	datastore      datastore.DataStore
}

func NewRuntime(output emitter.Emitter, dir string, name string, ds datastore.DataStore) *Runtime {
	return &Runtime{output: output, dir: dir, name: name, Status: "Starting", datastore: ds}
}

func (run *Runtime) NewTask(inputs map[string]interface{}) *Task {
	return &Task{Name: run.name, Runtime: run, Workdir: run.dir, Inputs: inputs, AllowLocalFiles: true, DataStore: run.datastore}
}

func (run *Runtime) Close() {
	log.Printf("Runtime closing")
	if run.output != nil {
		run.output.Close()
	}
	//run.man.DropRuntime(run.name)
}

func (run *Runtime) Emit(name string, o map[string]interface{}) error {
	return run.output.Emit(name, o)
}

func (run *Runtime) EmitObject(prefix string, c string, o map[string]interface{}) error {
	if prefix == "" {
		return run.output.EmitObject(run.name, c, o)
	}
	return run.output.EmitObject(prefix, c, o)
}

func (run *Runtime) EmitTable(prefix string, columns []string, sep rune) emitter.TableEmitter {
	return run.output.EmitTable(prefix, columns, sep)
}

func (run *Runtime) Printf(s string, x ...interface{}) {
	c := fmt.Sprintf(s, x...)
	log.Printf(c)
	run.Status = c
}

func (run *Runtime) GetCurrent() string {
	return run.Status
}

func (run *Runtime) GetStepNum() int64 {
	return run.StepCount
}

func (run *Runtime) GetStepTotal() int64 {
	return run.StepTotal
}

func (run *Runtime) SetStepCountTotal(i int64) {
	run.StepTotal = i
}

func (run *Runtime) AddStepCount(i int64) {
	atomic.AddInt64(&run.StepCount, i)
}

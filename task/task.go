package task

import (
	"io/ioutil"
	"path/filepath"

	"github.com/bmeg/sifter/loader"
)

type RuntimeTask interface {
	SetName(name string)
	SubTask(ext string) RuntimeTask

	Emit(name string, e map[string]interface{}) error

	GetInputs() map[string]interface{}
	AbsPath(p string) (string, error)
	TempDir() string
	WorkDir() string
	OutDir() string
	GetName() string

	Close()
}

type Task struct {
	Prefix  string
	Name    string
	Workdir string
	Inputs  map[string]interface{}
	Emitter loader.DataEmitter
	Outdir  string
}

func NewTask(name string, workDir string, outDir string, inputs map[string]interface{}) *Task {
	ld := loader.NewDirLoader(outDir)
	em, _ := ld.NewDataEmitter()

	workDir, _ = filepath.Abs(workDir)
	outDir, _ = filepath.Abs(outDir)

	return &Task{Name: name, Inputs: inputs, Workdir: workDir, Outdir: outDir, Emitter: em}
}

func (m *Task) SetName(name string) {
	m.Name = name
}

func (m *Task) Close() {
	m.Emitter.Close()
}

func (m *Task) GetName() string {
	if m.Prefix == "" {
		return m.Name
	}
	return m.Prefix + "." + m.Name
}

func (m *Task) SubTask(ext string) RuntimeTask {
	return &Task{
		Prefix:  m.GetName(),
		Name:    ext,
		Workdir: m.Workdir,
		Inputs:  m.Inputs,
		Emitter: m.Emitter,
		Outdir:  m.Outdir,
	}
}

func (m *Task) GetInputs() map[string]interface{} {
	return m.Inputs
}

func (m *Task) AbsPath(p string) (string, error) {
	//if !strings.HasPrefix(p, "/") {
	//	p = filepath.Join(m.Workdir, p)
	//}
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return a, nil
}

func (m *Task) OutDir() string {
	return m.Outdir
}

func (m *Task) TempDir() string {
	name, _ := ioutil.TempDir(m.Workdir, "tmp")
	return name
}

func (m *Task) WorkDir() string {
	return m.Workdir
}

func (m *Task) Emit(n string, e map[string]interface{}) error {
	return m.Emitter.Emit(m.GetName()+"."+n, e)
}

package task

import (
	"io/ioutil"
	"path/filepath"

	"github.com/bmeg/sifter/loader"
)

type RuntimeTask interface {
	Emit(name string, e map[string]interface{}) error

	GetInputs() map[string]interface{}
	AbsPath(p string) (string, error)
	TempDir() string
	WorkDir() string
	GetName() string
}

type Task struct {
	Name    string
	Workdir string
	Inputs  map[string]interface{}
	Emitter loader.DataEmitter
}

func (m *Task) GetName() string {
	return m.Name
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

func (m *Task) TempDir() string {
	name, _ := ioutil.TempDir(m.Workdir, "tmp")
	return name
}

func (m *Task) WorkDir() string {
	return m.Workdir
}

func (m *Task) Emit(n string, e map[string]interface{}) error {
	return m.Emitter.Emit(n, e)
}

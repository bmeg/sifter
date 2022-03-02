package task

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/loader"
)

type RuntimeTask interface {
	loader.DataEmitter

	GetInputs() map[string]interface{}
	AbsPath(p string) (string, error)
	TempDir() string
	WorkDir() string
	GetName() string
}

type Task struct {
	Name       string
	Workdir    string
	SourcePath string
	Inputs     map[string]interface{}
}

func (m *Task) GetName() string {
	return m.Name
}

func (m *Task) GetInputs() map[string]interface{} {
	return m.Inputs
}

func (m *Task) AbsPath(p string) (string, error) {
	if !strings.HasPrefix(p, "/") {
		p = filepath.Join(m.Workdir, p)
	}
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

package manager

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/loader"
)

type RuntimeTask interface {
	loader.DataEmitter

	GetInputs() map[string]interface{}
	Child(name string) RuntimeTask
	AbsPath(p string) (string, error)
	TempDir() string
	WorkDir() string
	GetName() string
	GetDataStore() (datastore.DataStore, error)
}

type Task struct {
	Name            string
	Runtime         *Runtime
	Workdir         string
	SourcePath      string
	Inputs          map[string]interface{}
	DataStore       datastore.DataStore
	AllowLocalFiles bool
}

func (m *Task) GetName() string {
	return m.Name
}

func (m *Task) GetInputs() map[string]interface{} {
	return m.Inputs
}

func (m *Task) Child(name string) RuntimeTask {
	cname := fmt.Sprintf("%s.%s", m.Name, name)
	return &Task{Name: cname, Runtime: m.Runtime, Workdir: m.Workdir, Inputs: m.Inputs, AllowLocalFiles: m.AllowLocalFiles, DataStore: m.DataStore}
}

func (m *Task) AbsPath(p string) (string, error) {
	if !strings.HasPrefix(p, "/") {
		p = filepath.Join(m.Workdir, p)
	}
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if !m.AllowLocalFiles {
		if !strings.HasPrefix(a, m.Workdir) {
			return "", fmt.Errorf("Input file not inside working directory")
		}
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

func (m *Task) Emit(name string, e map[string]interface{}) error {
	return m.Runtime.Emit(name, e)
}

func (m *Task) EmitObject(prefix string, c string, e map[string]interface{}) error {
	return m.Runtime.EmitObject(prefix, c, e)
}

func (m *Task) EmitTable(prefix string, columns []string, sep rune) loader.TableEmitter {
	return m.Runtime.EmitTable(prefix, columns, sep)
}

/*
func (m *Task) Output(name string, value string) error {
	if m.Runtime.OutputCallback != nil {
		return m.Runtime.OutputCallback(name, value)
	}
	return fmt.Errorf("Output Callback not set")
}
*/

func (m *Task) Printf(s string, x ...interface{}) {
	m.Runtime.Printf(s, x...)
}

func (m *Task) GetDataStore() (datastore.DataStore, error) {
	return m.DataStore, nil //DEBUG: fix this
}

func (m *Task) Close() {
	//m.Runtime.output.Close()
}

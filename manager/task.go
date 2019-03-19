package manager

import (
	"io/ioutil"
	"log"
	"path"

	"github.com/bmeg/grip/gripql"
	"github.com/hashicorp/go-getter"
)

type Task struct {
	Manager *Manager
	Workdir string
	Inputs  map[string]interface{}
}

func (man *Manager) NewTask(inputs map[string]interface{}) *Task {
	dir, err := ioutil.TempDir("./", "sifterwork_")
	if err != nil {
		log.Fatal(err)
	}
	return &Task{man, dir, inputs}
}

func (m *Task) Path(p string) string {
	return path.Join(m.Workdir, p)
}

func (m *Task) DownloadFile(url string) (string, error) {
	d := m.Path(path.Base(url))
	return d, getter.GetFile(d, url+"?archive=false")
}

func (m *Task) EmitVertex(v *gripql.Vertex) error {
	return m.Manager.EmitVertex(v)
}

func (m *Task) EmitEdge(e *gripql.Edge) error {
	return m.Manager.EmitEdge(e)
}

func (m *Task) Printf(s string, x ...interface{}) {
	m.Manager.Printf(s, x...)
}

package manager

import (
	"path"

	"github.com/bmeg/grip/gripql"
	"github.com/hashicorp/go-getter"
)

type Task struct {
	Manager *Manager
	Workdir string
	Inputs  map[string]interface{}
}

func (m *Task) Path(p string) string {
	return path.Join(m.Workdir, p)
}

func (m *Task) DownloadFile(url string, dest string) (string, error) {
	if dest == "" {
		dest = m.Path(path.Base(url))
	} else {
		dest = m.Path(dest)
	}
	return dest, getter.GetFile(dest, url+"?archive=false")
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

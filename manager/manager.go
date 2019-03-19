package manager

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/emitter"
)

type Manager struct {
	Playbooks   map[string]Playbook
	Output      emitter.Emitter
	Status      string
	VertexCount int64
	EdgeCount   int64
}

func Init(playbookDirs ...string) (*Manager, error) {
	pbMap := map[string]Playbook{}
	for _, pbDir := range playbookDirs {
		g, _ := filepath.Glob(filepath.Join(pbDir, "*.yaml"))
		for _, p := range g {
			pb := Playbook{}
			if err := ParseFile(p, &pb); err != nil {
				log.Printf("Parse Error: %s", err)
			} else {
				pbMap[pb.Name] = pb
			}
		}
	}

	//s := emitter.StdoutEmitter{}
	//s, _ := emitter.NewMongoEmitter("localhost:27017", "test")
	s, err := emitter.NewGripEmitter("localhost:8202", "test")
	return &Manager{pbMap, s, "Start", 0, 0}, err
}

func (m *Manager) Close() {

}

func (m *Manager) GetPlaybooks() []Playbook {
	out := make([]Playbook, 0, len(m.Playbooks))
	for _, i := range m.Playbooks {
		out = append(out, i)
	}
	return out
}

func (m *Manager) EmitVertex(v *gripql.Vertex) error {
	m.VertexCount += 1
	return m.Output.EmitVertex(v)
}

func (m *Manager) EmitEdge(e *gripql.Edge) error {
	m.EdgeCount += 1
	return m.Output.EmitEdge(e)
}

func (m *Manager) Printf(s string, x ...interface{}) {
	c := fmt.Sprintf(s, x...)
	log.Printf(c)
	m.Status = c
}

func (m *Manager) GetCurrent() string {
	return m.Status
}

func (m *Manager) GetVertexCount() int64 {
	return m.VertexCount
}

func (m Manager) GetEdgeCount() int64 {
	return m.EdgeCount
}

func (m Manager) GetStepNum() int64 {
	return 1
}

func (m Manager) GetStepTotal() int64 {
	return 10
}

package manager

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/bmeg/sifter/emitter"
)

type Manager struct {
	Playbooks   map[string]Playbook
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
	return &Manager{pbMap, "Start", 0, 0}, nil
}

func (m *Manager) Close() {
	//TODO: Cleanup the runtimes
}

func (m *Manager) NewEmitter(graph string) (emitter.Emitter, error) {
	//s := emitter.StdoutEmitter{}
	//s, _ := emitter.NewMongoEmitter("localhost:27017", "test")
	return emitter.NewGripEmitter("localhost:8202", graph)
}

func (m *Manager) GetPlaybooks() []Playbook {
	out := make([]Playbook, 0, len(m.Playbooks))
	for _, i := range m.Playbooks {
		out = append(out, i)
	}
	return out
}

func (m *Manager) GetPlaybook(name string) (Playbook, bool) {
	out, ok := m.Playbooks[name]
	return out, ok
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

func (m *Manager) GetEdgeCount() int64 {
	return m.EdgeCount
}

func (m *Manager) GetStepNum() int64 {
	return 1
}

func (m *Manager) GetStepTotal() int64 {
	return 10
}

func (m *Manager) NewRuntime(graph string) (Runtime, error) {
	dir, err := ioutil.TempDir("./", "sifterwork_")
	if err != nil {
		log.Fatal(err)
	}
	e, err := emitter.NewGripEmitter("localhost:8202", graph)
	return Runtime{m, e, dir}, err
}

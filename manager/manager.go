package manager

import (
	"github.com/bmeg/sifter/emitter"
	"github.com/bmeg/sifter/pipeline"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"
)

type Manager struct {
	Config    Config
	Playbooks map[string]Playbook
	Runtimes  sync.Map
	AllowLocalFiles bool
}

type Config struct {
	GripServer   string
	PlaybookDirs []string
	WorkDir      string
	ObjectOutput bool
}

func Init(config Config) (*Manager, error) {
	pbMap := map[string]Playbook{}
	for _, pbDir := range config.PlaybookDirs {
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
	return &Manager{config, pbMap, sync.Map{}, false}, nil
}

func (m *Manager) Close() {
	//TODO: Cleanup the runtimes
}

func (m *Manager) DropRuntime(name string) error {
	m.Runtimes.Delete(name)
	return nil
}

func (m *Manager) GraphExists(graph string) bool {
	o, err := emitter.GraphExists(m.Config.GripServer, graph)
	if err != nil {
		log.Printf("Failed to load graph driver: %s", err)
	}
	return o
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

func (m *Manager) NewRuntime(graph string) (*pipeline.Runtime, error) {
	dir, err := ioutil.TempDir(m.Config.WorkDir, "sifterwork_")
	if err != nil {
		log.Fatal(err)
	}
	dir, _ = filepath.Abs(dir)
	e, err := emitter.NewEmitter(m.Config.GripServer, graph)
	if err != nil {
		log.Printf("Emitter init failed: %s", err)
	}
	name := filepath.Base(dir)
	r := pipeline.NewRuntime(e, dir, name)
	m.Runtimes.Store(name, r)
	return r, err
}

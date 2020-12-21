package manager

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/pipeline"
	"github.com/bmeg/sifter/schema"
)

type Manager struct {
	Config          Config
	Playbooks       map[string]Playbook
	Runtimes        sync.Map
	AllowLocalFiles bool
	DataStore       datastore.DataStore
}

type Config struct {
	Driver       string
	PlaybookDirs []string
	WorkDir      string
	DataStore    *datastore.Config
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

	man := &Manager{config, pbMap, sync.Map{}, false, nil}

	if config.DataStore != nil {
		d, err := datastore.GetMongoStore(*config.DataStore)
		log.Printf("Mongo Error: %s", err)
		if err != nil {
			return nil, err
		}
		man.DataStore = d
	}

	return man, nil
}

func (m *Manager) Close() {
	//TODO: Cleanup the runtimes
}

func (m *Manager) DropRuntime(name string) error {
	m.Runtimes.Delete(name)
	return nil
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

func (m *Manager) NewRuntime(name string, dir string, sc *schema.Schemas) (*pipeline.Runtime, error) {
	dir, _ = filepath.Abs(dir)
	e, err := loader.NewDataEmitter(m.Config.Driver, sc)
	if err != nil {
		log.Printf("Emitter init failed: %s", err)
	}
	if name == "" {
		name = "default"
	}
	r := pipeline.NewRuntime(e, dir, name, m.DataStore)
	m.Runtimes.Store(name, r)
	return r, err
}

package manager

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/schema"
)

type Manager struct {
	Config    Config
	Runtimes  sync.Map
	DataStore datastore.DataStore
}

type Config struct {
	Loader    loader.Loader
	WorkDir   string
	DataStore *datastore.Config
}

func Init(config Config) (*Manager, error) {

	man := &Manager{config, sync.Map{}, nil}

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

func (m *Manager) NewRuntime(name string, dir string, sc *schema.Schemas) (*Runtime, error) {
	dir, _ = filepath.Abs(dir)
	e, err := m.Config.Loader.NewDataEmitter(sc)
	if err != nil {
		log.Printf("Emitter init failed: %s", err)
	}
	if name == "" {
		name = "default"
	}
	r := NewRuntime(e, dir, name, m.DataStore)
	m.Runtimes.Store(name, r)
	return r, err
}

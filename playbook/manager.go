package playbook

import (
	"sync"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/loader"
)

type Manager struct {
	Config   Config
	Runtimes sync.Map
}

type Config struct {
	Loader    loader.Loader
	WorkDir   string
	DataStore *datastore.Config
}

func Init(config Config) (*Manager, error) {

	man := &Manager{config, sync.Map{}}

	return man, nil
}

func (m *Manager) Close() {
	//TODO: Cleanup the runtimes
}

func (m *Manager) DropRuntime(name string) error {
	m.Runtimes.Delete(name)
	return nil
}

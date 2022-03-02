package playbook

import (
	"path/filepath"
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

func (m *Manager) NewRuntime(name string, dir string) (*Runtime, error) {
	dir, _ = filepath.Abs(dir)
	r := NewRuntime(dir, name)
	m.Runtimes.Store(name, r)
	return r, nil
}

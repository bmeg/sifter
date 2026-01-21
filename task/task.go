package task

import (
	"os"
	"path/filepath"

	"github.com/bmeg/sifter/loader"
)

type RuntimeTask interface {
	SetName(name string)

	Emit(name string, e map[string]interface{}) error

	GetConfig() map[string]string
	AbsPath(p string) (string, error)
	TempDir() string
	WorkDir() string
	BaseDir() string
	OutDir() string
	GetName() string

	Close()
}

type Task struct {
	Prefix  string
	Name    string
	Workdir string
	Basedir string
	Config  map[string]string
	Emitter loader.DataEmitter
	Outdir  string
}

func NewTask(name string, baseDir string, workDir string, outDir string, config map[string]string) *Task {
	ld := loader.NewDirLoader(outDir)
	em, _ := ld.NewDataEmitter()

	baseDir, _ = filepath.Abs(baseDir)
	workDir, _ = filepath.Abs(workDir)
	outDir, _ = filepath.Abs(outDir)

	return &Task{Name: name, Config: config, Basedir: baseDir, Workdir: workDir, Outdir: outDir, Emitter: em}
}

func (m *Task) SetName(name string) {
	m.Name = name
}

func (m *Task) Close() {
	m.Emitter.Close()
}

func (m *Task) GetName() string {
	if m.Prefix == "" {
		return m.Name
	}
	return m.Prefix + "." + m.Name
}

func (m *Task) GetConfig() map[string]string {
	return m.Config
}

func (m *Task) AbsPath(p string) (string, error) {
	//if !strings.HasPrefix(p, "/") {
	//	p = filepath.Join(m.Workdir, p)
	//}
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return a, nil
}

func (m *Task) OutDir() string {
	return m.Outdir
}

func (m *Task) TempDir() string {
	name, _ := os.MkdirTemp(m.Workdir, "tmp")
	return name
}

func (m *Task) WorkDir() string {
	return m.Workdir
}

func (m *Task) BaseDir() string {
	return m.Basedir
}

func (m *Task) Emit(name string, e map[string]interface{}) error {
	return m.Emitter.Emit(name, e)
}

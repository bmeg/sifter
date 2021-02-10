package test

import (
  "fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bmeg/sifter/cmd/run"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/manager"

  "path/filepath"

  "github.com/ghodss/yaml"
)

type ExampleConfig struct {
  Playbook string         `json:"playbook"`
  Inputs map[string]string `json:"inputs"`
  Outputs []string        `json:"outputs"`
}

func runPlaybook(playbook string, inputs map[string]interface{}, outdir string) error {
  workDir := "./"
  os.Mkdir(outdir, 0700)
	driver := "dir://" + outdir
	ld, err := loader.NewLoader(driver)
	if err != nil {
    return err
  }
	defer ld.Close()

	dir, err := ioutil.TempDir(workDir, "sifterwork_")
	defer os.RemoveAll(dir)

	man, err := manager.Init(manager.Config{Loader: ld, WorkDir: workDir, DataStore: nil})
	if err != nil {
    return err
	}
	defer man.Close()
	man.AllowLocalFiles = true

	err = run.Execute(playbook, dir, inputs, man)
	if err != nil {
		return err
	}
	os.RemoveAll(dir)
  return nil
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func TestExamples(t *testing.T) {
  tests, err := filepath.Glob("test-*.yaml")
  if err != nil {
    t.Error(err)
  }
  for _, tPath := range tests {
    raw, err := ioutil.ReadFile(tPath)
  	if err != nil {
  		t.Error(fmt.Errorf("failed to read config %s", tPath))
  	}
  	conf := ExampleConfig{}
  	if err := yaml.Unmarshal(raw, &conf); err != nil {
      t.Error(fmt.Errorf("failed to read config %s", tPath))
  	}
    inputs := map[string]interface{}{}
    for k, v := range conf.Inputs {
      inputs[k] = v
    }
    fmt.Printf("%s\n", conf)
    outDir, err := ioutil.TempDir("./", "testout_")
    runPlaybook(conf.Playbook, inputs, outDir)

    for _, out := range conf.Outputs {
      base := filepath.Base(out)
      dst := filepath.Join(outDir, base)
      if !fileExists(dst) {
        t.Errorf("Output %s not produced", base)
      }
    }
    os.RemoveAll(outDir)
  }
}

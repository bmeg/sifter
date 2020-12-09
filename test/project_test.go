package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/bmeg/sifter/manager"
  "github.com/bmeg/sifter/cmd/run"
)

func TestProject(t *testing.T) {
	workDir := "./"
	man, err := manager.Init(manager.Config{Driver: "dir://.", WorkDir: workDir, DataStore: nil})
	if err != nil {
		t.Error(err)
	}
	defer man.Close()
	man.AllowLocalFiles = true
	inputs := map[string]interface{}{}
	dir, err := ioutil.TempDir(workDir, "sifterwork_")
  defer os.RemoveAll(dir)

	err = run.Execute("./resources/project.yaml", dir, inputs, man)
  if err != nil {
		t.Error(err)
	}
	os.RemoveAll(dir)
}

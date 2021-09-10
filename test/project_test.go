package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/bmeg/sifter/cmd/run"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/manager"
)

func TestProject(t *testing.T) {
	workDir := "./"
	driver := "dir://."
	ld, err := loader.NewLoader(driver)
	if err != nil {
		t.Error(err)
	}
	defer ld.Close()

	dir, err := ioutil.TempDir(workDir, "sifterwork_")
	defer os.RemoveAll(dir)

	man, err := manager.Init(manager.Config{Loader: ld, WorkDir: workDir, DataStore: nil})
	if err != nil {
		t.Error(err)
	}
	defer man.Close()
	inputs := map[string]interface{}{}

	err = run.Execute("./resources/project.yaml", dir, "./", inputs, man)
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(dir)
}

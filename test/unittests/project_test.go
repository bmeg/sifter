package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/bmeg/sifter/cmd/run"
	"github.com/bmeg/sifter/loader"
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

	if err != nil {
		t.Error(err)
	}
	inputs := map[string]string{}

	err = run.Execute("../resources/project.yaml", dir, "./", inputs)
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(dir)
}

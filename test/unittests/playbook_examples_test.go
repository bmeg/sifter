package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/bmeg/sifter/cmd/run"
	"github.com/bmeg/sifter/loader"

	"path/filepath"

	"github.com/ghodss/yaml"
)

type PlaybookExampleConfig struct {
	Playbook string            `json:"playbook"`
	Inputs   map[string]string `json:"inputs"`
	Outputs  []string          `json:"outputs"`
}

func runPlaybook(playbook string, inputs map[string]string, outdir string) error {
	workDir := "./"
	os.Mkdir(outdir, 0700)
	driver := "dir://" + outdir
	ld, err := loader.NewLoader(driver)
	if err != nil {
		return err
	}
	defer ld.Close()

	dir, _ := os.MkdirTemp(workDir, "sifterwork_")
	defer os.RemoveAll(dir)

	err = run.ExecuteFile(playbook, dir, outdir, inputs)
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

func TestPlaybookExamples(t *testing.T) {
	tests, err := filepath.Glob("test-playbook-*.yaml")
	if err != nil {
		t.Error(err)
	}
	for _, tPath := range tests {
		raw, err := os.ReadFile(tPath)
		if err != nil {
			t.Error(fmt.Errorf("failed to read config %s", tPath))
		}
		conf := PlaybookExampleConfig{}
		if err := yaml.Unmarshal(raw, &conf); err != nil {
			t.Error(fmt.Errorf("failed to read config %s", tPath))
		}
		inputs := map[string]string{}
		for k, v := range conf.Inputs {
			inputs[k] = v
		}
		fmt.Printf("%s\n", conf)
		outDir, _ := os.MkdirTemp("./", "testout_")
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

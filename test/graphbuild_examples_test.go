package test

import (
  "testing"
  "path/filepath"
  "io/ioutil"
  "fmt"
  "github.com/ghodss/yaml"
  "os"
  "github.com/bmeg/sifter/cmd/graph"
)

type GraphBuildExampleConfig struct {
	GraphMap string            `json:"graphMap"`
	InputDir string            `json:"inputDir"`
	Outputs  []string          `json:"outputs"`
}


func TestGraphBuildExamples(t *testing.T) {
  tests, err := filepath.Glob("test-graphbuild-*.yaml")
  if err != nil {
    t.Error(err)
  }
  for _, tPath := range tests {
    raw, err := ioutil.ReadFile(tPath)
    if err != nil {
      t.Error(fmt.Errorf("failed to read config %s", tPath))
    }
    conf := GraphBuildExampleConfig{}
    if err := yaml.Unmarshal(raw, &conf); err != nil {
      t.Error(fmt.Errorf("failed to read config %s", tPath))
    }
    tmpDir, err := ioutil.TempDir("./", "sifterwork_")
    outDir, err := ioutil.TempDir("./", "outdir_")

    err = graph.RunGraphBuild(conf.GraphMap, conf.InputDir, tmpDir, outDir)

    for _, out := range conf.Outputs {
			base := filepath.Base(out)
			dst := filepath.Join(outDir, base)
			if !fileExists(dst) {
				t.Errorf("Output %s not produced", base)
			}
		}

    os.RemoveAll(tmpDir)
    os.RemoveAll(outDir)
    if err != nil {
      t.Error(err)
    }
  }
}

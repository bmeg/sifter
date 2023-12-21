package test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

var tPath string = "config.yaml"

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

type CommandLineConfig struct {
	LineCount []int    `json:"linecount"`
	Playbook  string   `json:"playbook"`
	Outputs   []string `json:"outputs"`
}

/*
This test checks three things:
1. The file exists in the expected place in config.yaml
2. The file contains the expercted number of data lines
3. The file's name is the expected name. Tests useName option added in this PR
*/
func TestCommandLines(t *testing.T) {
	raw, err := ioutil.ReadFile(tPath)
	if err != nil {
		t.Error(fmt.Errorf("failed to read config %s %s", tPath, err))
	}
	conf := []CommandLineConfig{}
	if err := yaml.UnmarshalStrict(raw, &conf); err != nil {
		t.Error(fmt.Errorf("failed to read config %s %s", tPath, err))
	}
	// read in conf, ie config.yaml in this case
	for _, c := range conf {
		cmd := exec.Command("../sifter", "run", c.Playbook)
		fmt.Printf("Running: %s\n", c.Playbook)
		err = cmd.Run()
		if err != nil {
			t.Errorf("Failed running %s: %s", c.Playbook, err)
		} else {
			for i, chk := range c.Outputs {
				// iterate through expected output paths
				path := filepath.Join(filepath.Dir(c.Playbook), chk)
				fmt.Printf("Checking %s \n", path)
				if stat, err := os.Stat(path); err == nil {
					if stat.Size() > 0 {
						file, err := os.Open(path)
						var reader io.Reader
						if strings.HasSuffix(path, ".gz") {
							reader, _ = gzip.NewReader(file)
						} else {
							reader = file
						}
						defer file.Close()
						if err == nil {
							// count of data lines of each expected output file and compare to expected value in config
							count, _ := lineCounter(reader)

							if count != c.LineCount[i] {
								t.Errorf("Incorrect number of lines: %d != %d", count, c.LineCount[i])
							}
						} else {
							t.Errorf("failed to open %s", path)
						}
					}
				} else if os.IsNotExist(err) {
					t.Errorf("Output %s missing", path)
				}

				// remove test files when done testing
				os.RemoveAll(path)
			}
		}

	}
}

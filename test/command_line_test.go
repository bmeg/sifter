package test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
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
	Playbook string   `json:"playbook"`
	Checks   []string `json:"checks"`
	Outputs  []string `json:"outputs"`
}

func TestCommandLines(t *testing.T) {
	raw, err := ioutil.ReadFile(tPath)
	if err != nil {
		t.Error(fmt.Errorf("failed to read config %s %s", tPath, err))
	}
	conf := []CommandLineConfig{}
	if err := yaml.UnmarshalStrict(raw, &conf); err != nil {
		t.Error(fmt.Errorf("failed to read config %s %s", tPath, err))
	}

	for _, c := range conf {
		cmd := exec.Command("../sifter", "run", c.Playbook)
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stderr
		fmt.Printf("Running: %s\n", c.Playbook)
		err = cmd.Run()
		if err != nil {
			t.Errorf("Failed running %s", c.Playbook)
		} else {
			/*
				for f, chk := range c.Outputs {
					path := filepath.Join("out", f)
					if _, err := os.Stat(path); err == nil {
						if chk.LineCount != 0 {
							file, err := os.Open(path)
							var reader io.Reader
							if strings.HasSuffix(path, ".gz") {
								reader, _ = gzip.NewReader(file)
							} else {
								reader = file
							}
							defer file.Close()
							if err == nil {
								count, _ := lineCounter(reader)
								if count != chk.LineCount {
									t.Errorf("Incorrect number of lines: %d != %d", count, chk.LineCount)
								}
							} else {
								t.Errorf("failed to open %s", f)
							}
						}
					} else if os.IsNotExist(err) {
						t.Errorf("Output %s missing", f)
					}
				}
			*/
		}
		os.RemoveAll("./out")
	}
}

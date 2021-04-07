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

	"github.com/ghodss/yaml"
	shellquote "github.com/kballard/go-shellquote"
)

var tPath string = "command-line-tests.yaml"

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

type OutputConfig struct {
	LineCount int `json:"lineCount"`
}

type CommandLineConfig struct {
	Title   string                  `json:"title"`
	Command string                  `json:"command"`
	Outputs map[string]OutputConfig `json:"outputs"`
}

func TestCommandLines(t *testing.T) {
	raw, err := ioutil.ReadFile(tPath)
	if err != nil {
		t.Error(fmt.Errorf("failed to read config %s %s", tPath, err))
	}
	conf := []CommandLineConfig{}
	if err := yaml.Unmarshal(raw, &conf); err != nil {
		t.Error(fmt.Errorf("failed to read config %s %s", tPath, err))
	}

	for _, c := range conf {
		csplit, err := shellquote.Split(c.Command)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("%#v\n", csplit)
		cmd := exec.Command("../sifter", csplit...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("Running: %#v", cmd)
		err = cmd.Run()
		if err != nil {
			t.Error(err)
		} else {
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
		}
		os.RemoveAll("./out")
	}
}

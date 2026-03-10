package test

import (
	"bufio"
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
		t.Logf("Running: %s\n", c.Playbook)
		err = cmd.Run()
		if err != nil {
			t.Errorf("Failed running %s: %s", c.Playbook, err)
		} else {
			for i, chk := range c.Outputs {
				// iterate through expected output paths
				path := filepath.Join(filepath.Dir(c.Playbook), chk)
				t.Logf("Checking %s \n", path)
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

// TestCaptureMode verifies that --capture-dir and --capture-limit flags create
// NDJSON capture files and respect the record limit.
func TestCaptureMode(t *testing.T) {
	captureDir, err := os.MkdirTemp("", "sifter-capture-*")
	if err != nil {
		t.Fatalf("Failed to create temp capture dir: %s", err)
	}
	defer os.RemoveAll(captureDir)

	playbook := "examples/gene-table/gene-table.yaml"
	limit := 3

	cmd := exec.Command("../sifter", "run",
		"--capture-dir", captureDir,
		"--capture-limit", fmt.Sprintf("%d", limit),
		playbook,
	)
	t.Logf("Running: %s with capture-dir=%s capture-limit=%d", playbook, captureDir, limit)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed running %s: %s", playbook, err)
	}

	// Check that at least one .ndjson file was created
	entries, err := os.ReadDir(captureDir)
	if err != nil {
		t.Fatalf("Failed to read capture dir: %s", err)
	}
	if len(entries) == 0 {
		t.Errorf("Expected capture NDJSON files in %s, but directory is empty", captureDir)
		return
	}

	// Verify each capture file has at most `limit` records
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ndjson") {
			continue
		}
		filePath := filepath.Join(captureDir, entry.Name())
		f, err := os.Open(filePath)
		if err != nil {
			t.Errorf("Failed to open capture file %s: %s", filePath, err)
			continue
		}

		lineCount := 0
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if scanner.Text() != "" {
				lineCount++
			}
		}
		scanErr := scanner.Err()
		f.Close()

		if scanErr != nil {
			t.Errorf("Error reading capture file %s: %s", filePath, scanErr)
		}
		if lineCount > limit {
			t.Errorf("Capture file %s has %d records, expected at most %d", entry.Name(), lineCount, limit)
		}
		t.Logf("Capture file %s has %d records (limit=%d)", entry.Name(), lineCount, limit)
	}

	// Clean up the playbook output
	outputDir := filepath.Join(filepath.Dir(playbook), "output")
	os.RemoveAll(outputDir)
}

package extractors

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"

	xj "github.com/basgys/goxml2json"
)

type XMLLoadStep struct {
	Input string `json:"input"`
}

func (ml *XMLLoadStep) Start(task task.RuntimeTask) (chan map[string]any, error) {
	//log.Printf("Starting XML Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return nil, fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", input)

	fhd, err := os.Open(input)
	if err != nil {
		return nil, err
	}

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return nil, err
		}
	} else {
		hd = fhd
	}

	procChan := make(chan map[string]any, 100)

	go func() {
		jStr, err := xj.Convert(hd)
		if err == nil {
			data := map[string]any{}
			if err = json.Unmarshal(jStr.Bytes(), &data); err == nil {
				procChan <- data
			}
		}
		close(procChan)
	}()
	return procChan, nil
}

func (ml *XMLLoadStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(ml.Input) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

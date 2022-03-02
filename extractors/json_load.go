package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	"github.com/bmeg/sifter/transform"
)

type JSONLoadStep struct {
	Input     string         `json:"input" jsonschema_description:"Path of multiline JSON file to transform"`
	Transform transform.Pipe `json:"transform" jsonschema_description:"Transformation Pipeline"`
	Multiline bool           `json:"multiline" jsonschema_description:"Load file as a single multiline JSON object"`
}

func (ml *JSONLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	log.Printf("Starting JSON Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetInputs(), nil)
	inputPath, err := task.AbsPath(input)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)

	var reader chan []byte
	if ml.Multiline {
		reader = make(chan []byte, 1)
		dat, err := os.ReadFile(inputPath)
		if err != nil {
			return nil, err
		}
		reader <- dat
		close(reader)
	} else {
		if strings.HasSuffix(inputPath, ".gz") {
			reader, err = golib.ReadGzipLines(inputPath)
		} else {
			reader, err = golib.ReadFileLines(inputPath)
		}
		if err != nil {
			return nil, err
		}
	}

	procChan := make(chan map[string]interface{}, 100)
	go func() {
		for line := range reader {
			o := map[string]interface{}{}
			if len(line) > 0 {
				json.Unmarshal(line, &o)
				procChan <- o
			}
		}
		close(procChan)
	}()
	return procChan, nil
}

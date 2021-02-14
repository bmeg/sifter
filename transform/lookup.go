package transform

import (
	"fmt"
	"log"
	"os"

	"strings"

	"encoding/json"

	"github.com/bmeg/golib"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/pipeline"
)

type JSONFileLookupStep struct {
	Input   string `json:"input"`
	Field   string `json:"field"`
	Key     string `json:"key"`
	Project map[string]string
	Copy    map[string]string
	//found it more space efficiant to store the JSON rather then keep
	//all the unpacked values
	table map[string][]byte //map[string]interface{}
}

func (jf *JSONFileLookupStep) Init(task *pipeline.Task) error {
	input, err := evaluate.ExpressionString(jf.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)
	if err != nil {
		return err
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading Translation file: %s", inputPath)

	var inputStream chan []byte
	if strings.HasSuffix(inputPath, ".gz") {
		inputStream, err = golib.ReadGzipLines(inputPath)
	} else {
		inputStream, err = golib.ReadFileLines(inputPath)
	}
	if err != nil {
		return err
	}

	jf.table = map[string][]byte{} //map[string]interface{}{}

	for line := range inputStream {
		if len(line) > 0 {
			row := map[string]interface{}{}
			err := json.Unmarshal(line, &row)
			if err != nil {
				return err
			}
			if key, ok := row[jf.Key]; ok {
				if keyStr, ok := key.(string); ok {
					jf.table[keyStr] = line
				}
			}
		}
	}
	return nil
}

func (jf *JSONFileLookupStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	field, err := evaluate.ExpressionString(jf.Field, task.Inputs, i)
	if err == nil {
		if line, ok := jf.table[field]; ok {
			row := map[string]interface{}{}
			json.Unmarshal(line, &row)
			for k, v := range jf.Copy {
				if ki, ok := row[v]; ok {
					i[k] = ki
				}
			}
			for k, v := range jf.Project {
				val, err := evaluate.ExpressionString(v, task.Inputs, row)
				if err == nil {
					err = SetProjectValue(i, k, val)
					if err != nil {
						log.Printf("project error: %s", err)
					}
				}
			}
		}
	}
	return i
}

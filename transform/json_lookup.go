package transform

import (
	"fmt"
	"log"
	"os"

	"strings"

	"encoding/json"

	"github.com/bmeg/golib"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type JSONFileLookupStep struct {
	Input   string `json:"input"`
	Value   string `json:"value"`
	Key     string `json:"key"`
	Project map[string]string
	Copy    map[string]string
	Replace *TableReplace
}

type jsonLookupProcess struct {
	config *JSONFileLookupStep
	inputs map[string]any
	table  map[string][]byte //found it more space efficiant to store the JSON rather then keep all the unpacked values
}

func (jf *JSONFileLookupStep) Init(task task.RuntimeTask) (Processor, error) {
	inputPath, err := evaluate.ExpressionString(jf.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	if s, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("File Not Found: %s", inputPath)
	} else if s.IsDir() {
		return nil, fmt.Errorf("File Not Found: %s", inputPath)
	}
	log.Printf("Loading Translation file: %s", inputPath)

	var inputStream chan []byte
	if strings.HasSuffix(inputPath, ".gz") {
		inputStream, err = golib.ReadGzipLines(inputPath)
	} else {
		inputStream, err = golib.ReadFileLines(inputPath)
	}
	if err != nil {
		return nil, err
	}

	//finish this
	//do a table based value replacement
	if jf.Replace != nil {
		table := map[string]string{}
		for line := range inputStream {
			if len(line) > 0 {
				row := map[string]interface{}{}
				err := json.Unmarshal(line, &row)
				if err != nil {
					return nil, err
				}
				if key, ok := row[jf.Key]; ok {
					if keyStr, ok := key.(string); ok {
						table[keyStr] = "" // row[jf.]
					}
				}
			}
		}
		return &tableReplaceInst{jf.Replace, table}, nil
	}

	jp := &jsonLookupProcess{jf, task.GetConfig(), map[string][]byte{}}
	for line := range inputStream {
		if len(line) > 0 {
			row := map[string]interface{}{}
			err := json.Unmarshal(line, &row)
			if err != nil {
				return nil, err
			}
			if key, ok := row[jf.Key]; ok {
				if keyStr, ok := key.(string); ok {
					jp.table[keyStr] = line
				}
			}
		}
	}
	log.Printf("jsonLookup loaded %d values from %s", len(jp.table), inputPath)

	return jp, nil
}

func (jp *jsonLookupProcess) Close() {}

func (jp *jsonLookupProcess) Process(i map[string]interface{}) []map[string]interface{} {
	field, err := evaluate.ExpressionString(jp.config.Value, jp.inputs, i)
	if err == nil {
		if line, ok := jp.table[field]; ok {
			row := map[string]interface{}{}
			json.Unmarshal(line, &row)
			for k, v := range jp.config.Copy {
				if ki, ok := row[v]; ok {
					i[k] = ki
				}
			}
			for k, v := range jp.config.Project {
				val, err := evaluate.ExpressionString(v, jp.inputs, row)
				if err == nil {
					err = setProjectValue(i, k, val)
					if err != nil {
						log.Printf("project error: %s", err)
					}
				}
			}
		}
	}
	return []map[string]any{i}
}

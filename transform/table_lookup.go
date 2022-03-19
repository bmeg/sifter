package transform

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/task"

	"github.com/bmeg/sifter/evaluate"
)

type TableLookupStep struct {
	Input   string   `json:"input"`
	Sep     string   `json:"sep"`
	Value   string   `json:"value"`
	Key     string   `json:"key"`
	Header  []string `json:"header"`
	Project map[string]string
}

type tableLookupProcess struct {
	config *TableLookupStep
	inputs map[string]any
	colmap map[string]int
	table  map[string][]string
}

func (tr *TableLookupStep) Init(task task.RuntimeTask) (Processor, error) {
	inputPath, err := evaluate.ExpressionString(tr.Input, task.GetConfig(), nil)
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

	if tr.Sep == "" {
		tr.Sep = "\t"
	}

	tp := &tableLookupProcess{config: tr, inputs: task.GetConfig()}

	tp.colmap = nil
	if len(tr.Header) > 0 {
		tp.colmap = map[string]int{}
		for i, n := range tr.Header {
			tp.colmap[n] = i
		}
	}
	tp.table = map[string][]string{}
	for line := range inputStream {
		if len(line) > 0 {
			row := strings.Split(string(line), tr.Sep)
			if tp.colmap == nil {
				tp.colmap = map[string]int{}
				for i, k := range row {
					tp.colmap[k] = i
				}
			} else {
				tp.table[row[tp.colmap[tr.Key]]] = row
			}
		}
	}
	log.Printf("tableLookup loaded %d values from %s", len(tp.table), inputPath)
	return tp, nil
}

func (tp *tableLookupProcess) Close() {}

func (tp *tableLookupProcess) Process(i map[string]interface{}) []map[string]interface{} {
	field, err := evaluate.ExpressionString(tp.config.Value, tp.inputs, i)
	if err == nil {
		if pv, ok := tp.table[field]; ok {
			for k, v := range tp.config.Project {
				if ki, ok := tp.colmap[v]; ok {
					i[k] = pv[ki]
				}
			}
		}
	}
	return []map[string]any{i}
}

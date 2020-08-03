package transform

import (
	"fmt"
	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/emitter"
	"log"
	"os"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/pipeline"
)

type TableWriteStep struct {
	Output  string   `json:"output" jsonschema_description:"Name of file to create"`
	Columns []string `json:"columns" jsonschema_description:"Columns to be written into table file"`
	Sep     string `json:"sep"`
	emit    emitter.TableEmitter
}

type TableReplaceStep struct {
	Input string `json:"input"`
	Field string `json:"field"`
	table map[string]string
}

type TableProjectStep struct {
	Input   string `json:"input"`
  Sep     string `json:"sep"`
	Field   string `json:"field"`
	Project map[string]string
	header  map[string]int
	table   map[string][]string
}

func (tw *TableWriteStep) Init(task *pipeline.Task) {
	sep := '\t'
	if tw.Sep != "" {
		sep = rune(tw.Sep[0])
	}
	tw.emit = task.Runtime.EmitTable(tw.Output, tw.Columns, sep)
}

func (tw *TableWriteStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	if err := tw.emit.EmitRow(i); err != nil {
		log.Printf("Row Error: %s", err)
	}
	return i
}

func (tw *TableWriteStep) Close() {
	log.Printf("Closing tableWriter: %s", tw.Output)
	tw.emit.Close()
}

func (tr *TableReplaceStep) Init(task *pipeline.Task) error {
	input, err := evaluate.ExpressionString(tr.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)
	if err != nil {
		return err
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)

	inputStream, err := golib.ReadFileLines(inputPath)
	if err != nil {
		return err
	}
	tr.table = map[string]string{}
	for line := range inputStream {
		if len(line) > 0 {
			row := strings.Split(string(line), "\t")
			tr.table[row[0]] = row[1]
		}
	}
	return nil
}

func (tw *TableReplaceStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {

	if _, ok := i[tw.Field]; ok {
		out := map[string]interface{}{}
		for k, v := range i {
			if k == tw.Field {
				if x, ok := v.(string); ok {
					if n, ok := tw.table[x]; ok {
						out[k] = n
					} else {
						out[k] = x
					}
				} else if x, ok := v.([]interface{}); ok {
					o := []interface{}{}
					for _, y := range x {
						if z, ok := y.(string); ok {
							if n, ok := tw.table[z]; ok {
								o = append(o, n)
							} else {
								o = append(o, z)
							}
						}
					}
					out[k] = o
				} else {
					out[k] = v
				}
			} else {
				out[k] = v
			}
		}
		return out
	}
	return i
}

func (tr *TableProjectStep) Init(task *pipeline.Task) error {
	input, err := evaluate.ExpressionString(tr.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)
	if err != nil {
		return err
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading Translation file: %s", inputPath)

	inputStream, err := golib.ReadFileLines(inputPath)
	if err != nil {
		return err
	}
  if tr.Sep == "" {
    tr.Sep = "\t"
  }
	tr.header = nil
	tr.table = map[string][]string{}
	for line := range inputStream {
		if len(line) > 0 {
			row := strings.Split(string(line), tr.Sep)
			if tr.header == nil {
        tr.header = map[string]int{}
        for i, k := range row {
          tr.header[k] = i
        }
			} else {
				tr.table[row[0]] = row
			}
		}
	}
	return nil
}

func (tw *TableProjectStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	if fv, ok := i[tw.Field]; ok {
    if fstr, ok := fv.(string); ok {
      if pv, ok := tw.table[fstr]; ok {
  		  for k, v := range tw.Project {
          if ki, ok := tw.header[v]; ok {
            i[k] = pv[ki]
          }
  			}
  		}
    }
	}
  return i
}

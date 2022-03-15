package transform

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type TableReplaceStep struct {
	Input  string            `json:"input"`
	Table  map[string]string `json:"table"`
	Field  string            `json:"field"`
	Target string            `json:"target"`
	table  map[string]string
}

func (tr *TableReplaceStep) Init(task task.RuntimeTask) error {
	if tr.Input != "" {
		input, err := evaluate.ExpressionString(tr.Input, task.GetInputs(), nil)
		inputPath, err := task.AbsPath(input)
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
	} else if len(tr.Table) > 0 {
		tr.table = map[string]string{}
		for k, v := range tr.Table {
			tr.table[k] = v
		}
	}
	return nil
}

func (tr *TableReplaceStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {

	if _, ok := i[tr.Field]; ok {
		out := map[string]interface{}{}
		for k, v := range i {
			if k == tr.Field {
				d := k
				if tr.Target != "" {
					d = tr.Target
				}
				if x, ok := v.(string); ok {
					if n, ok := tr.table[x]; ok {
						out[d] = n
					} else {
						out[d] = x
					}
				} else if x, ok := v.([]interface{}); ok {
					o := []interface{}{}
					for _, y := range x {
						if z, ok := y.(string); ok {
							if n, ok := tr.table[z]; ok {
								o = append(o, n)
							} else {
								o = append(o, z)
							}
						}
					}
					out[d] = o
				} else {
					out[d] = v
				}
			} else {
				out[k] = v
			}
		}
		return out
	}
	return i
}

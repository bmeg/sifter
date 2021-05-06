package transform

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/loader"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/manager"
)

type TableWriteStep struct {
	Output  string   `json:"output" jsonschema_description:"Name of file to create"`
	Columns []string `json:"columns" jsonschema_description:"Columns to be written into table file"`
	Sep     string   `json:"sep"`
	emit    loader.TableEmitter
}

type TableReplaceStep struct {
	Input  string            `json:"input"`
	Table  map[string]string `json:"table"`
	Field  string            `json:"field"`
	Target string            `json:"target"`
	table  map[string]string
}

type TableLookupStep struct {
	Input   string   `json:"input"`
	Sep     string   `json:"sep"`
	Field   string   `json:"field"`
	Key     string   `json:"key"`
	Header  []string `json:"header"`
	Project map[string]string
	colmap  map[string]int
	table   map[string][]string
}

func (tw *TableWriteStep) Init(task manager.RuntimeTask) {
	sep := '\t'
	if tw.Sep != "" {
		sep = rune(tw.Sep[0])
	}
	tw.emit = task.EmitTable(tw.Output, tw.Columns, sep)
}

func (tw *TableWriteStep) Run(i map[string]interface{}, task manager.RuntimeTask) map[string]interface{} {
	if err := tw.emit.EmitRow(i); err != nil {
		log.Printf("Row Error: %s", err)
	}
	return i
}

func (tw *TableWriteStep) Close() {
	log.Printf("Closing tableWriter: %s", tw.Output)
	tw.emit.Close()
}

func (tr *TableReplaceStep) Init(task manager.RuntimeTask) error {
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

func (tr *TableReplaceStep) Run(i map[string]interface{}, task manager.RuntimeTask) map[string]interface{} {

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

func (tr *TableLookupStep) Init(task manager.RuntimeTask) error {
	input, err := evaluate.ExpressionString(tr.Input, task.GetInputs(), nil)
	inputPath, err := task.AbsPath(input)
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

	if tr.Sep == "" {
		tr.Sep = "\t"
	}
	tr.colmap = nil
	if len(tr.Header) > 0 {
		tr.colmap = map[string]int{}
		for i, n := range tr.Header {
			tr.colmap[n] = i
		}
	}
	tr.table = map[string][]string{}
	for line := range inputStream {
		if len(line) > 0 {
			row := strings.Split(string(line), tr.Sep)
			if tr.colmap == nil {
				tr.colmap = map[string]int{}
				for i, k := range row {
					tr.colmap[k] = i
				}
			} else {
				tr.table[row[tr.colmap[tr.Key]]] = row
			}
		}
	}
	return nil
}

func (tr *TableLookupStep) Run(i map[string]interface{}, task manager.RuntimeTask) map[string]interface{} {
	field, err := evaluate.ExpressionString(tr.Field, task.GetInputs(), i)
	if err == nil {
		if pv, ok := tr.table[field]; ok {
			for k, v := range tr.Project {
				if ki, ok := tr.colmap[v]; ok {
					i[k] = pv[ki]
				}
			}
		}
	}
	return i
}

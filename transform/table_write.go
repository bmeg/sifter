package transform

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type TableWriteStep struct {
	Output           string   `json:"output" jsonschema_description:"Name of file to create"`
	Columns          []string `json:"columns" jsonschema_description:"Columns to be written into table file"`
	Header           string   `json:"header"`
	SkipColumnHeader bool     `json:"skipColumnHeader"`
	Sep              string   `json:"sep"`
}

type tableWriteProcess struct {
	config  *TableWriteStep
	columns []string
	out     io.WriteCloser
	handle  io.WriteCloser
	writer  *csv.Writer
}

func (tw *TableWriteStep) Init(task task.RuntimeTask) (Processor, error) {
	sep := '\t'
	if tw.Sep != "" {
		sep = rune(tw.Sep[0])
	}

	output, err := evaluate.ExpressionString(tw.Output, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}
	outputPath := filepath.Join(task.OutDir(), output)
	log.Printf("Starting TableWriter to %s", outputPath)

	te := tableWriteProcess{}
	te.handle, _ = os.Create(outputPath)
	if strings.HasSuffix(outputPath, ".gz") {
		te.out = gzip.NewWriter(te.handle)
	} else {
		te.out = te.handle
	}
	if tw.Header != "" {
		te.out.Write([]byte(tw.Header))
		te.out.Write([]byte("\n"))
	}
	te.writer = csv.NewWriter(te.out)
	te.writer.Comma = sep
	te.columns = tw.Columns
	te.config = tw
	if !tw.SkipColumnHeader {
		te.writer.Write(te.columns)
	}
	return &te, nil
}

func (tw *TableWriteStep) GetOutputs(task task.RuntimeTask) []string {
	output, err := evaluate.ExpressionString(tw.Output, task.GetConfig(), nil)
	if err != nil {
		return []string{}
	}
	outputPath := filepath.Join(task.OutDir(), output)
	log.Printf("table output %s %s", task.OutDir(), output)
	return []string{outputPath}
}

func (tp *tableWriteProcess) PoolReady() bool {
	return false
}

func (tp *tableWriteProcess) Process(i map[string]any) map[string]any {
	o := make([]string, len(tp.columns))
	for j, k := range tp.columns {
		if v, ok := i[k]; ok {
			if vStr, ok := v.(string); ok {
				o[j] = vStr
			} else {
				b, _ := json.Marshal(v)
				o[j] = string(b)
			}
		}
	}
	tp.writer.Write(o)
	return i
}

func (tp *tableWriteProcess) Close() {
	log.Printf("Closing tableWriter: %s", tp.config.Output)
	tp.writer.Flush()
	tp.out.Close()
	tp.handle.Close()
}

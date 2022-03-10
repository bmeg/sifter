package writers

import (
	"compress/gzip"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type TableWriter struct {
	Output  string   `json:"output" jsonschema_description:"Name of file to create"`
	Columns []string `json:"columns" jsonschema_description:"Columns to be written into table file"`
	Sep     string   `json:"sep"`
}

type tableWriteProcess struct {
	config  *TableWriter
	sep     rune
	columns []string
	out     io.WriteCloser
	handle  io.WriteCloser
	writer  *csv.Writer
}

func (tw *TableWriter) Init(task task.RuntimeTask) (WriteProcess, error) {
	log.Printf("Starting TableWriter")
	sep := '\t'
	if tw.Sep != "" {
		sep = rune(tw.Sep[0])
	}

	output, err := evaluate.ExpressionString(tw.Output, task.GetInputs(), nil)
	if err != nil {
		return nil, err
	}
	outputPath, err := task.AbsPath(output)

	te := tableWriteProcess{}
	te.handle, _ = os.Create(outputPath)
	if strings.HasSuffix(outputPath, ".gz") {
		te.out = gzip.NewWriter(te.handle)
	} else {
		te.out = te.handle
	}
	te.writer = csv.NewWriter(te.out)
	te.writer.Comma = sep
	te.columns = tw.Columns
	te.writer.Write(te.columns)
	return &te, nil
}

func (tw *TableWriter) GetOutputs(task task.RuntimeTask) []string {
	output, err := evaluate.ExpressionString(tw.Output, task.GetInputs(), nil)
	if err != nil {
		return []string{}
	}
	outputPath, err := task.AbsPath(output)
	return []string{outputPath}
}

func (tp *tableWriteProcess) Write(i map[string]interface{}) {
	o := make([]string, len(tp.columns))
	for j, k := range tp.columns {
		if v, ok := i[k]; ok {
			if vStr, ok := v.(string); ok {
				o[j] = vStr
			}
		}
	}
	tp.writer.Write(o)
}

func (tp *tableWriteProcess) Close() {
	log.Printf("Closing tableWriter: %s", tp.config.Output)
	tp.writer.Flush()
	tp.out.Close()
	tp.handle.Close()
}
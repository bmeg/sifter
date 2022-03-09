package writers

import (
	"log"

	"github.com/bmeg/sifter/task"
)

type TableWriter struct {
	Output  string   `json:"output" jsonschema_description:"Name of file to create"`
	Columns []string `json:"columns" jsonschema_description:"Columns to be written into table file"`
	Sep     string   `json:"sep"`
}

type tableWriteProcess struct {
	config *TableWriter
	sep    rune
}

func (tw *TableWriter) Init(task task.RuntimeTask) (WriteProcess, error) {
	log.Printf("Starting TableWriter")
	sep := '\t'
	if tw.Sep != "" {
		sep = rune(tw.Sep[0])
	}
	return &tableWriteProcess{tw, sep}, nil
}

func (tp *tableWriteProcess) Write(i map[string]interface{}) {
	log.Printf("TSV Writer: %#v", i)
}

func (tp *tableWriteProcess) Close() {
	log.Printf("Closing tableWriter: %s", tp.config.Output)
	//tw.emit.Close()
}

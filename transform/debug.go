package transform

import (
	"encoding/json"
	"log"

	"github.com/bmeg/sifter/task"
)

type DebugStep struct {
	Label  string `json:"label"`
	Format bool   `json:"format"`
}

func (ds DebugStep) Init(task task.RuntimeTask) (Processor, error) {
	return ds, nil
}

func (ds DebugStep) Process(i map[string]interface{}) []map[string]interface{} {
	var s []byte
	if ds.Format {
		s, _ = json.MarshalIndent(i, "", "    ")
	} else {
		s, _ = json.Marshal(i)
	}
	log.Printf("DebugData %s: %s", ds.Label, s)
	return []map[string]any{i}
}

func (ds DebugStep) Close() {}

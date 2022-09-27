package transform

import (
	"encoding/json"
	"log"

	"github.com/bmeg/sifter/task"
)

type DebugStep struct {
	Label string `json:"label"`
}

func (ds DebugStep) Init(task task.RuntimeTask) (Processor, error) {
	return ds, nil
}

func (db DebugStep) Process(i map[string]interface{}) []map[string]interface{} {
	s, _ := json.Marshal(i)
	log.Printf("DebugData %s: %s", db.Label, s)
	return []map[string]any{i}
}

func (db DebugStep) Close() {}

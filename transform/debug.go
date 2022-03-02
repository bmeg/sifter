package transform

import (
	"encoding/json"
	"log"

	"github.com/bmeg/sifter/task"
)

type DebugStep struct {
	Label string `json:"label"`
}

func (db DebugStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {
	s, _ := json.Marshal(i)
	log.Printf("DebugData %s: %s", db.Label, s)
	return i
}

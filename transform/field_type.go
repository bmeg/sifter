package transform

import (
	"strconv"

	"github.com/bmeg/sifter/task"
)

type FieldTypeStep map[string]string

type fieldTypeProcess struct {
	config FieldTypeStep
}

func (fs FieldTypeStep) Init(task task.RuntimeTask) (Processor, error) {
	return &fieldTypeProcess{fs}, nil
}

func (fp *fieldTypeProcess) Close() {}

func (fp *fieldTypeProcess) Process(row map[string]interface{}) []map[string]interface{} {
	o := map[string]interface{}{}
	for x, y := range row {
		o[x] = y
	}
	for field, fType := range fp.config {
		if fType == "int" || fType == "integer" {
			if val, ok := row[field]; ok {
				if vStr, ok := val.(string); ok {
					if d, err := strconv.ParseInt(vStr, 10, 64); err == nil {
						o[field] = d
					} else {
						o[field] = nil
					}
				}
			}
		} else if fType == "float" {
			if val, ok := row[field]; ok {
				if vStr, ok := val.(string); ok {
					if d, err := strconv.ParseFloat(vStr, 64); err == nil {
						o[field] = d
					} else {
						o[field] = nil
					}
				}
			}
		} else if fType == "list" {
			if val, ok := o[field]; ok {
				switch val.(type) {
				case []string:
				case []interface{}:
				case []float64:
				case []int:
				default:
					o[field] = []interface{}{val}
				}
			}
		}
	}
	return []map[string]any{o}
}

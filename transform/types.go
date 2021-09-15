package transform

import (
	"strconv"

	"github.com/bmeg/sifter/manager"
)

type FieldTypeStep map[string]string

func (fs FieldTypeStep) Run(i map[string]interface{}, task manager.RuntimeTask) map[string]interface{} {
	o := map[string]interface{}{}
	for x, y := range i {
		o[x] = y
	}
	for field, fType := range fs {
		if fType == "int" {
			if val, ok := i[field]; ok {
				if vStr, ok := val.(string); ok {
					if d, err := strconv.ParseInt(vStr, 10, 64); err == nil {
						o[field] = d
					}
				}
			}
		} else if fType == "float" {
			if val, ok := i[field]; ok {
				if vStr, ok := val.(string); ok {
					if d, err := strconv.ParseFloat(vStr, 64); err == nil {
						o[field] = d
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
	return o
}

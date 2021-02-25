package transform

import (
	"strconv"

	"github.com/bmeg/sifter/manager"
)

type FieldTypeStep map[string]string

func (fs FieldTypeStep) Run(i map[string]interface{}, task *manager.Task) map[string]interface{} {
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
		}
	}
	return o
}

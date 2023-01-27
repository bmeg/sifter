package schema

import (
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func (s GraphSchema) Validate(classID string, data map[string]any) error {
	class := s.GetClass(classID)
	if class != nil {
		return class.Validate(data)
	}
	return fmt.Errorf("class '%s' not found", classID)
}

func (s GraphSchema) CleanAndValidate(class *jsonschema.Schema, data map[string]any) (map[string]any, error) {
	out := map[string]any{}
	for k, v := range data {
		if _, ok := class.Properties[k]; ok {
			out[k] = v
		}
	}
	return out, class.Validate(out)
}

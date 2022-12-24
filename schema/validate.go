package schema

import "fmt"

func (s GraphSchema) Validate(classID string, data map[string]any) error {
	if class, ok := s.Classes[classID]; ok {
		return class.Validate(data)
	}
	return fmt.Errorf("class '%s' not found", classID)
}

func (s GraphSchema) CleanAndValidate(classID string, data map[string]any) (map[string]any, error) {
	if class, ok := s.Classes[classID]; ok {
		out := map[string]any{}
		for k, v := range data {
			if _, ok := class.Properties[k]; ok {
				out[k] = v
			}
		}
		return out, class.Validate(out)
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}

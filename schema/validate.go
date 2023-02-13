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
		if subCls, ok := class.Properties[k]; ok {
			if isObjectSchema(subCls) {
				if vMap, ok := v.(map[string]any); ok {
					vn, err := s.CleanAndValidate(subCls, vMap)
					if err == nil {
						out[k] = vn
					} else {
						return nil, err
					}
				}
			} else if isArraySchema(subCls) && isObjectSchema(subCls.Items2020) {
				if vArray, ok := v.([]any); ok {
					o := []any{}
					for _, v := range vArray {
						if vMap, ok := v.(map[string]any); ok {
							l, err := s.CleanAndValidate(subCls.Items2020, vMap)
							if err == nil {
								o = append(o, l)
							} else {
								return nil, err
							}
						}
					}
					out[k] = o
				}
			} else {
				out[k] = v
			}
		} else {
			if class.AdditionalProperties != nil {
				if addParam, ok := class.AdditionalProperties.(bool); ok {
					if addParam {
						out[k] = v
					}
				} else if addParam, ok := class.AdditionalProperties.(*jsonschema.Schema); ok {
					if err := addParam.Validate(v); err == nil {
						out[k] = v
					}
				}
			}
		}
	}
	return out, class.Validate(out)
}

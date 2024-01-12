package transform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CodeBlock struct {
	Code    string
	Ref     string
	BaseDir string
}

func (cb *CodeBlock) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &cb.Code); err == nil {
		return nil
	}
	ref := map[string]any{}
	if err := json.Unmarshal(data, &ref); err == nil {
		if path, ok := ref["$ref"]; ok {
			if pathStr, ok := path.(string); ok {
				cb.Ref = pathStr
				return nil
			}
		}
	}
	return fmt.Errorf("unknown code block type")
}

func (cb *CodeBlock) SetBaseDir(path string) {
	cb.BaseDir = path
}

func (cb *CodeBlock) String() string {
	if cb.Ref != "" {
		path := filepath.Join(cb.BaseDir, cb.Ref)
		data, err := os.ReadFile(path)
		if err == nil {
			cb.Code = string(data)
		}
	}
	return cb.Code
}

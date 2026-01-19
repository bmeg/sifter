package config

import "strings"

type Params map[string]Param

type Param struct {
	Type    string `json:"type"`
	Default any    `json:"default,omitempty"`
}

type ParamRequest struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type Configurable interface {
	GetRequiredParams() []ParamRequest
}

func (in *Param) IsFile() bool {
	return strings.ToLower(in.Type) == "file"
}

func (in *Param) IsDir() bool {
	return strings.ToLower(in.Type) == "path"
}

func TrimPrefix(s string) string {
	if strings.HasPrefix(s, "params.") {
		return strings.TrimPrefix(s, "params.")
	}
	return s
}

func (in *ParamRequest) IsFile() bool {
	return strings.ToLower(in.Type) == "file"
}

func (in *ParamRequest) IsDir() bool {
	return strings.ToLower(in.Type) == "path"
}

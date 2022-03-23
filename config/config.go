package config

import "strings"

type ConfigVar struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default string `json:"default"`
}

type Config map[string]ConfigVar

type Configurable interface {
	GetConfigFields() []ConfigVar
}

func (in *ConfigVar) IsFile() bool {
	if in.Type == "File" || in.Type == "file" {
		return true
	}
	return false
}

func (in *ConfigVar) IsDir() bool {
	if in.Type == "Dir" || in.Type == "dir" || in.Type == "Directory" || in.Type == "directory" {
		return true
	}
	return false
}

func TrimPrefix(s string) string {
	return strings.TrimPrefix(s, "config.")
}

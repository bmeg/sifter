package config

import "strings"

type Config map[string]string

type Type string

const (
	Unknown Type = ""
	File    Type = "File"
	Dir     Type = "Dir"
)

type ConfigVar struct {
	Name string `json:"name"`
	Type Type
}

type Configurable interface {
	GetConfigFields() []ConfigVar
}

func (in *ConfigVar) IsFile() bool {
	return in.Type == File
}

func (in *ConfigVar) IsDir() bool {
	return in.Type == Dir
}

func TrimPrefix(s string) string {
	return strings.TrimPrefix(s, "config.")
}

package config

import "strings"

type Config map[string]*string

type Type string

const (
	Unknown Type = ""
	File    Type = "File"
	Dir     Type = "Dir"
)

type Variable struct {
	Name string `json:"name"`
	Type Type
}

type Configurable interface {
	GetConfigFields() []Variable
}

func (in *Variable) IsFile() bool {
	return in.Type == File
}

func (in *Variable) IsDir() bool {
	return in.Type == Dir
}

func TrimPrefix(s string) string {
	return strings.TrimPrefix(s, "config.")
}

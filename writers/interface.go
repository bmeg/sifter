package writers

import (
	"fmt"
	"reflect"

	"github.com/bmeg/sifter/task"
)

type WriteProcess interface {
	Write(map[string]any)
	Close()
}

type Writer interface {
	From() string
	Init(task.RuntimeTask) (WriteProcess, error)
	GetOutputs(task.RuntimeTask) []string
}

type WriteConfig struct {
	TableWrite     *TableWriter     `json:"tableWrite"`
	SnakeFileWrite *SnakeFileWriter `json:"snakefile"`
}

func (wc *WriteConfig) From() string {
	v := reflect.ValueOf(wc).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		x := f.Interface()
		if z, ok := x.(Writer); ok {
			if !f.IsNil() {
				return z.From()
			}
		}
	}
	return ""
}

func (wc *WriteConfig) Init(t task.RuntimeTask) (WriteProcess, error) {
	v := reflect.ValueOf(wc).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		x := f.Interface()
		if z, ok := x.(Writer); ok {
			if !f.IsNil() {
				return z.Init(t)
			}
		}
	}
	return nil, fmt.Errorf(("Writer not defined"))
}

func (wc *WriteConfig) GetOutputs(t task.RuntimeTask) []string {
	v := reflect.ValueOf(wc).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		x := f.Interface()
		if z, ok := x.(Writer); ok {
			if !f.IsNil() {
				return z.GetOutputs(t)
			}
		}
	}
	return []string{}
}

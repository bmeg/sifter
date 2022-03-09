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
	Init(task.RuntimeTask) (WriteProcess, error)
}

type WriteConfig struct {
	TableWriter *TableWriter `json:"tableWrite"`
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

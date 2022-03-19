package transform

import (
	"github.com/bmeg/sifter/task"
)

type FromStep string

func (f FromStep) Init(t task.RuntimeTask) (Processor, error) {
	return f, nil
}

func (f FromStep) Process(i map[string]any) []map[string]any {
	return []map[string]any{i}
}

func (f FromStep) Close() {}

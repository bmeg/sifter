package extractors

import "github.com/bmeg/sifter/task"

type EmbeddedLoader []map[string]any

func (el *EmbeddedLoader) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	out := make(chan map[string]any, 10)
	go func() {
		defer close(out)
		for _, row := range *el {
			out <- row
		}
	}()
	return out, nil
}

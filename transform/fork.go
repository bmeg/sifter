package transform

import (
	"sync"

	"github.com/bmeg/sifter/manager"
)

type ForkStep struct {
	Transform []Pipe `json:"transform"`
}

func (fs *ForkStep) Init(task manager.RuntimeTask) error {
	for _, t := range fs.Transform {
		t.Init(task)
	}
	return nil
}

func (fs *ForkStep) Start(in chan map[string]interface{}, task manager.RuntimeTask, wg *sync.WaitGroup) (chan map[string]interface{}, error) {

	out := make(chan map[string]interface{}, 10)

	inchan := []chan map[string]interface{}{}
	touts := []chan map[string]interface{}{}
	for _, t := range fs.Transform {
		i := make(chan map[string]interface{}, 10)
		inchan = append(inchan, i)
		o, _ := t.Start(i, task.Child("fork"), wg)
		touts = append(touts, o)
		go func() {
			//Filter does not emit the output of its sub pipeline, but it has to digest it
			for range o {
			}
		}()
	}

	go func() {
		//Filter emits a copy of its input, without changing it
		defer close(out)
		for i := range in {
			out <- i //copy of input is passed along to output unchanged
			for _, ic := range inchan {
				ic <- i
			}
		}
		for _, ic := range inchan {
			close(ic)
		}
	}()
	return out, nil
}

func (fs *ForkStep) Close() {
	for _, t := range fs.Transform {
		t.Close()
	}
}

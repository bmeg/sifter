package transform

import (
	"log"
	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type FilterStep struct {
	Field   string `json:"field"`
	Match   string `json:"match"`
	Check   string `json:"check" jsonschema_description:"How to check value, 'exists' or 'hasValue'"`
	Method  string `json:"method"`
	Python  string `json:"python"`
	GPython string `json:"gpython"`
	Steps   Pipe   `json:"steps"`
	inChan  chan map[string]interface{}
	proc    evaluate.Processor
}

func (fs *FilterStep) Init(task task.RuntimeTask) {
	if fs.Python != "" && fs.Method != "" {
		log.Printf("Starting Filter Map: %s", fs.Python)
		e := evaluate.GetEngine("python", task.WorkDir())
		c, err := e.Compile(fs.Python, fs.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		fs.proc = c
	} else if fs.GPython != "" && fs.Method != "" {
		log.Printf("Starting Filter Map: %s", fs.GPython)
		e := evaluate.GetEngine("gpython", task.WorkDir())
		c, err := e.Compile(fs.GPython, fs.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		fs.proc = c
	}
	fs.Steps.Init(task)
}

func (fs FilterStep) Start(in chan map[string]interface{}, task task.RuntimeTask, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	out := make(chan map[string]interface{}, 10)
	fs.inChan = make(chan map[string]interface{}, 100)
	tout, _ := fs.Steps.Start(fs.inChan, task.Child("filter"), wg)
	go func() {
		//Filter does not emit the output of its sub pipeline, but it has to digest it
		for range tout {
		}
	}()

	go func() {
		//Filter emits a copy of its input, without changing it
		defer close(out)
		defer close(fs.inChan)
		for i := range in {
			fs.run(i, task)
			out <- i
		}
	}()
	return out, nil
}

func (fs FilterStep) run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {
	if fs.proc != nil {
		out, err := fs.proc.EvaluateBool(i)
		if err != nil {
			log.Printf("Filter Error: %s", err)
		}
		if out {
			fs.inChan <- i
		}
		return i
	}
	col, err := evaluate.ExpressionString(fs.Field, task.GetInputs(), i)
	if (fs.Check == "" && fs.Match == "") || fs.Check == "exists" {
		if err == nil {
			fs.inChan <- i
		}
		return i
	} else if fs.Check == "hasValue" {
		if err == nil && col != "" {
			fs.inChan <- i
		}
		return i
	}

	match, _ := evaluate.ExpressionString(fs.Match, task.GetInputs(), i)
	if col == match {
		fs.inChan <- i
	}
	return i
}

func (fs FilterStep) Close() {
	if fs.proc != nil {
		fs.proc.Close()
	}
	fs.Steps.Close()
}

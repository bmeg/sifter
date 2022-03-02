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
}

func (fs FilterStep) Start(in chan map[string]interface{}, task task.RuntimeTask, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	out := make(chan map[string]interface{}, 10)

	go func() {
		//Filter emits a copy of its input, without changing it
		defer close(out)
		for i := range in {
			if fs.run(i, task) {
				out <- i
			}
		}
	}()
	return out, nil
}

func (fs FilterStep) run(i map[string]interface{}, task task.RuntimeTask) bool {
	if fs.proc != nil {
		out, err := fs.proc.EvaluateBool(i)
		if err != nil {
			log.Printf("Filter Error: %s", err)
		}
		return out
	}
	col, err := evaluate.ExpressionString(fs.Field, task.GetInputs(), i)
	if (fs.Check == "" && fs.Match == "") || fs.Check == "exists" {
		if err == nil {
			return true
		}
		return false
	} else if fs.Check == "hasValue" {
		if err == nil && col != "" {
			return true
		}
		return false
	}

	match, _ := evaluate.ExpressionString(fs.Match, task.GetInputs(), i)
	if col == match {
		return true
	}
	return false
}

func (fs FilterStep) Close() {
	if fs.proc != nil {
		fs.proc.Close()
	}
}

package transform

import (

	//"sync"
	"log"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type MapStep struct {
	Method  string `json:"method" jsonschema_description:"Name of function to call"`
	Python  string `json:"python" jsonschema_description:"Python code to be run"`
	GPython string `json:"gpython" jsonschema_description:"Python code to be run using GPython"`
	proc    evaluate.Processor
}

func (ms *MapStep) Init(task task.RuntimeTask) {
	if ms.Python != "" {
		log.Printf("Init Map: %s", ms.Python)
		e := evaluate.GetEngine("python", task.WorkDir())
		c, err := e.Compile(ms.Python, ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		ms.proc = c
	} else if ms.GPython != "" {
		log.Printf("Init Map: %s", ms.GPython)
		e := evaluate.GetEngine("gpython", task.WorkDir())
		c, err := e.Compile(ms.GPython, ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		ms.proc = c
	}
}

func (ms *MapStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {
	out, err := ms.proc.Evaluate(i)
	if err != nil {
		log.Printf("Map Step error: %s", err)
	}
	return out
}

func (ms *MapStep) Close() {
	ms.proc.Close()
}

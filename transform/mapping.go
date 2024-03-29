package transform

import (

	//"sync"
	"fmt"
	"log"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type MapStep struct {
	Method  string     `json:"method" jsonschema_description:"Name of function to call"`
	Python  string     `json:"python" jsonschema_description:"Python code to be run"`
	GPython *CodeBlock `json:"gpython" jsonschema_description:"Python code to be run using GPython"`
}

type mapProcess struct {
	config *MapStep
	proc   evaluate.Processor
}

func (ms *MapStep) Init(task task.RuntimeTask) (Processor, error) {
	if ms.Python != "" {
		log.Printf("Init Map: %s", ms.Python)
		e := evaluate.GetEngine("python", task.WorkDir())
		c, err := e.Compile(ms.Python, ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		return &mapProcess{ms, c}, nil
	} else if ms.GPython != nil {
		log.Printf("Init Map: %s", ms.GPython)
		ms.GPython.SetBaseDir(task.BaseDir())
		e := evaluate.GetEngine("gpython", task.WorkDir())
		c, err := e.Compile(ms.GPython.String(), ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		return &mapProcess{ms, c}, nil
	}
	return nil, fmt.Errorf("script not found")
}

func (mp *mapProcess) PoolReady() bool {
	return true
}

func (mp *mapProcess) Process(i map[string]interface{}) map[string]interface{} {
	out, err := mp.proc.Evaluate(i)
	if err != nil {
		log.Printf("Map Step error: %s", err)
	}
	return out
}

func (mp *mapProcess) Close() {
	mp.proc.Close()
}

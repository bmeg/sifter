package transform

import (

	//"sync"
	"fmt"
	"log"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type FlatMapStep struct {
	Method  string `json:"method" jsonschema_description:"Name of function to call"`
	Python  string `json:"python" jsonschema_description:"Python code to be run"`
	GPython string `json:"gpython" jsonschema_description:"Python code to be run using GPython"`
}

type flatMapProcess struct {
	config *FlatMapStep
	proc   evaluate.Processor
}

func (ms *FlatMapStep) Init(task task.RuntimeTask) (Processor, error) {
	if ms.Python != "" {
		log.Printf("Init Map: %s", ms.Python)
		e := evaluate.GetEngine("python", task.WorkDir())
		c, err := e.Compile(ms.Python, ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		return &flatMapProcess{ms, c}, nil
	} else if ms.GPython != "" {
		log.Printf("Init Map: %s", ms.GPython)
		e := evaluate.GetEngine("gpython", task.WorkDir())
		c, err := e.Compile(ms.GPython, ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		return &flatMapProcess{ms, c}, nil
	}
	return nil, fmt.Errorf("script not found")
}

func (mp *flatMapProcess) PoolReady() bool {
	return true
}

func (mp *flatMapProcess) Process(i map[string]interface{}) []map[string]interface{} {
	o, err := mp.proc.EvaluateArray(i)
	if err != nil {
		log.Printf("FlatMap Step error: %s", err)
	}
	//log.Printf("Flatmap out: %#v", o)
	out := []map[string]any{}
	for _, i := range o {
		if m, ok := i.(map[string]any); ok {
			out = append(out, m)
		} else {
			log.Printf("Flatmap output error: %#v", i)
		}
	}
	return out
}

func (mp *flatMapProcess) Close() {
	mp.proc.Close()
}

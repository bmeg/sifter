package transform

import (

	//"sync"
	"fmt"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type FlatMapStep struct {
	Method  string     `json:"method" jsonschema_description:"Name of function to call"`
	Python  string     `json:"python" jsonschema_description:"Python code to be run"`
	GPython *CodeBlock `json:"gpython" jsonschema_description:"Python code to be run using GPython"`
}

type flatMapProcess struct {
	config *FlatMapStep
	proc   evaluate.Processor
}

func (ms *FlatMapStep) Init(task task.RuntimeTask) (Processor, error) {
	if ms.Python != "" {
		logger.Debug("Init Map: %s", ms.Python)
		e := evaluate.GetEngine("python", task.WorkDir())
		c, err := e.Compile(ms.Python, ms.Method)
		if err != nil {
			logger.Error("Compile Error: %s", err)
		}
		return &flatMapProcess{ms, c}, nil
	} else if ms.GPython != nil {
		logger.Debug("Init Map: %s", ms.GPython)
		ms.GPython.SetBaseDir(task.BaseDir())
		e := evaluate.GetEngine("gpython", task.WorkDir())
		c, err := e.Compile(ms.GPython.String(), ms.Method)
		if err != nil {
			logger.Error("Compile Error: %s", err)
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
		logger.Error("FlatMap Step error: %s", err)
	}
	out := []map[string]any{}
	for _, i := range o {
		if m, ok := i.(map[string]any); ok {
			out = append(out, m)
		} else {
			logger.Error("Flatmap output error: %#v", i)
		}
	}
	return out
}

func (mp *flatMapProcess) Close() {
	mp.proc.Close()
}

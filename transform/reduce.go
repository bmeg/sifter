package transform

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type ReduceStep struct {
	Field    string                  `json:"field"`
	Method   string                  `json:"method"`
	Python   string                  `json:"python"`
	GPython  *CodeBlock              `json:"gpython"`
	InitData *map[string]interface{} `json:"init"`
}

type reduceProcess struct {
	config *ReduceStep
	proc   evaluate.Processor
}

func (ms *ReduceStep) Init(t task.RuntimeTask) (Processor, error) {
	if ms.Python != "" {
		log.Printf("ReduceInit: %s", ms.InitData)
		log.Printf("Reduce: %s", ms.Python)
		e := evaluate.GetEngine("python", t.WorkDir())
		c, err := e.Compile(ms.Python, ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		return &reduceProcess{ms, c}, nil
	} else if ms.GPython != nil {
		ms.GPython.SetBaseDir(t.BaseDir())
		log.Printf("ReduceInit: %s", ms.InitData)
		log.Printf("Reduce: %s", ms.GPython)
		e := evaluate.GetEngine("gpython", t.WorkDir())
		c, err := e.Compile(ms.GPython.String(), ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		return &reduceProcess{ms, c}, nil
	}
	return nil, fmt.Errorf("script not found")
}

func (rp *reduceProcess) Close() {
	rp.proc.Close()
}

func (rp *reduceProcess) GetInit() map[string]any {
	if rp.config.InitData == nil {
		return map[string]any{}
	}
	return *rp.config.InitData
}

func (rp *reduceProcess) GetKey(i map[string]any) string {
	if x, ok := i[rp.config.Field]; ok {
		if xStr, ok := x.(string); ok {
			return xStr
		}
	} else {
		log.Printf("Missing field in reduce")
	}
	return ""
}

func (rp *reduceProcess) Reduce(key string, a map[string]any, b map[string]any) map[string]any {
	out, err := rp.proc.Evaluate(a, b)
	if err != nil {
		log.Printf("Reduce Error: %s", err)
	}
	return out
}

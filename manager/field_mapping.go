package manager


import (
  "sync"
  "log"
  "github.com/bmeg/sifter/evaluate"
)

type MapStep struct {
  Method string `json:"method"`
  Python string `json:"python"`
  pyCode *evaluate.PyCode
}


func (ms *MapStep) Start(task *Task, wg *sync.WaitGroup) {
  log.Printf("Starting Map: %s", ms.Python)
  c, err := evaluate.PyCompile(ms.Python)
  if err != nil {
    log.Printf("%s", err)
  }
  ms.pyCode = c
}

func (ms *MapStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  out := ms.pyCode.Evaluate(ms.Method, i)
  return out
}

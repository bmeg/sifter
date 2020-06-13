
package steps

import (
  "os"
  "log"
  "fmt"
  "sync"
  "strings"
  "encoding/json"
  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/transform"
  "github.com/bmeg/sifter/pipeline"
  "github.com/bmeg/golib"
)

type JSONLoadStep struct {
  Input         string                    `json:"input" jsonschema_description:"Path of multiline JSON file to transform"`
  Transform     transform.TransformPipe   `json:"transform" jsonschema_description:"Transformation Pipeline"`
  SkipIfMissing bool                      `json:"skipIfMissing" jsonschema_description:"Skip without error if file does note exist"`
}

func (ml *JSONLoadStep) Run(task *pipeline.Task) error {
  log.Printf("Starting JSON Load")
	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		if ml.SkipIfMissing {
			return nil
		}
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)

  var reader chan []byte
	if strings.HasSuffix(inputPath, ".gz") {
		reader, err = golib.ReadGzipLines(inputPath)
	} else {
		reader, err = golib.ReadFileLines(inputPath)
	}
  if err != nil {
    return err
  }
  procChan := make(chan map[string]interface{}, 100)
  wg := &sync.WaitGroup{}

  if err := ml.Transform.Init( task ); err != nil {
    return err
  }

  out, err := ml.Transform.Start( procChan, task, wg )
  if err != nil {
    return err
  }
  go func() {
    for range out {}
  }()

  for line := range reader {
    o := map[string]interface{}{}
    if len(line) > 0 {
      json.Unmarshal(line, &o)
      procChan <- o
    }
  }

  log.Printf("Done Loading")
  close(procChan)
  wg.Wait()

	return nil
}

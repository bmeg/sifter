package extractors

import (
	"fmt"
	"log"
	"os"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"

	"github.com/bmeg/sifter/transform"

	"github.com/linkedin/goavro/v2"
)

type AvroLoadStep struct {
	Input     string         `json:"input" jsonschema_description:"Path of avro object file to transform"`
	Transform transform.Pipe `json:"transform" jsonschema_description:"Transformation Pipeline"`
}

func (ml *AvroLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	log.Printf("Starting Avro Load")

	input, err := evaluate.ExpressionString(ml.Input, task.GetInputs(), nil)

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return nil, fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", input)

	fh, err := os.Open(input)
	if err != nil {
		return nil, err
	}

	ocf, err := goavro.NewOCFReader(fh)
	if err != nil {
		return nil, err
	}

	procChan := make(chan map[string]interface{}, 100)
	go func() {
		for ocf.Scan() {
			datum, err := ocf.Read()
			if err == nil {
				if d, ok := datum.(map[string]interface{}); ok {
					procChan <- d
				}
			}
		}
		close(procChan)
		fh.Close()
		log.Printf("Done Loading")
	}()

	return procChan, nil
}

package extractors

import (
	"fmt"
	"os"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"

	"github.com/linkedin/goavro/v2"
)

type AvroLoadStep struct {
	Input string `json:"input" jsonschema_description:"Path of avro object file to transform"`
}

func (ml *AvroLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	logger.Debug("Starting Avro Load")

	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", input)
	}
	logger.Debug("Loading: %s", input)

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
		logger.Debug("Done Loading")
	}()

	return procChan, nil
}

func (ml *AvroLoadStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(ml.Input) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

package extractors

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/transform"

	"github.com/linkedin/goavro/v2"
)

type AvroLoadStep struct {
	Input         string         `json:"input" jsonschema_description:"Path of avro object file to transform"`
	Transform     transform.Pipe `json:"transform" jsonschema_description:"Transformation Pipeline"`
	SkipIfMissing bool           `json:"skipIfMissing" jsonschema_description:"Skip without error if file does note exist"`
}

func (ml *AvroLoadStep) Run(task *manager.Task) error {
	log.Printf("Starting Avro Load")

	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
	inputPath, err := task.AbsPath(input)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		if ml.SkipIfMissing {
			return nil
		}
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)

	fh, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	ocf, err := goavro.NewOCFReader(fh)
	if err != nil {
		return err
	}

	procChan := make(chan map[string]interface{}, 100)
	wg := &sync.WaitGroup{}

	if err := ml.Transform.Init(task); err != nil {
		return err
	}

	out, err := ml.Transform.Start(procChan, task, wg)
	if err != nil {
		return err
	}
	go func() {
		for range out {
		}
	}()

	for ocf.Scan() {
		datum, err := ocf.Read()
		if err == nil {
			if d, ok := datum.(map[string]interface{}); ok {
				procChan <- d
			}
		}
	}

	log.Printf("Done Loading")
	close(procChan)
	wg.Wait()
	ml.Transform.Close()
	fh.Close()
	return nil
}

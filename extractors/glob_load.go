package extractors

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type GlobLoadStep struct {
	Input     string         `json:"input" jsonschema_description:"Path of avro object file to transform"`
	XMLLoad   *XMLLoadStep   `json:"xmlLoad"`
	TableLoad *TableLoadStep `json:"tableLoad" jsonschema_description:"Run transform pipeline on a TSV or CSV"`
	JSONLoad  *JSONLoadStep  `json:"jsonLoad" jsonschema_description:"Run a transform pipeline on a multi line json file"`
	AvroLoad  *AvroLoadStep  `json:"avroLoad" jsonschema_description:"Load data from avro file"`
}

func (gl *GlobLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	log.Printf("Starting Glob Load")
	input, err := evaluate.ExpressionString(gl.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}
	if gl.XMLLoad != nil {
		flist, err := filepath.Glob(input)
		if err != nil {
			return nil, err
		}
		out := make(chan map[string]any, 10)
		go func() {
			defer close(out)
			for _, f := range flist {
				a := *gl.XMLLoad
				a.Input = f
				o, err := a.Start(task)
				if err == nil {
					for i := range o {
						out <- i
					}
				}
			}
		}()
		return out, nil
	} else if gl.JSONLoad != nil {
		flist, err := filepath.Glob(input)
		if err != nil {
			return nil, err
		}
		out := make(chan map[string]any, 10)
		go func() {
			defer close(out)
			for _, f := range flist {
				a := *gl.JSONLoad
				a.Input = f
				o, err := a.Start(task)
				if err == nil {
					for i := range o {
						out <- i
					}
				}
			}
		}()
		return out, nil
	}
	return nil, fmt.Errorf("Not found")
}

func (gl *GlobLoadStep) GetConfigFields() []config.ConfigVar {
	out := []config.ConfigVar{}
	for _, s := range evaluate.ExpressionIDs(gl.Input) {
		out = append(out, config.ConfigVar{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

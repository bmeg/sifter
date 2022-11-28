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
	StoreFilename string         `json:"storeFilename"`
	Input         string         `json:"input" jsonschema_description:"Path of avro object file to transform"`
	XMLLoad       *XMLLoadStep   `json:"xmlLoad"`
	TableLoad     *TableLoadStep `json:"tableLoad" jsonschema_description:"Run transform pipeline on a TSV or CSV"`
	JSONLoad      *JSONLoadStep  `json:"jsonLoad" jsonschema_description:"Run a transform pipeline on a multi line json file"`
	AvroLoad      *AvroLoadStep  `json:"avroLoad" jsonschema_description:"Load data from avro file"`
}

func (gl *GlobLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	log.Printf("Starting Glob Load")
	input, err := evaluate.ExpressionString(gl.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}
	if gl.XMLLoad != nil || gl.JSONLoad != nil || gl.TableLoad != nil {
		flist, err := filepath.Glob(input)
		if err != nil {
			return nil, err
		}
		out := make(chan map[string]any, 10)
		go func() {
			defer close(out)
			for count, f := range flist {
				log.Printf("Glob %d of %d", count, len(flist))
				//var a func()
				var a Source
				if gl.XMLLoad != nil {
					t := *gl.XMLLoad
					t.Input = f
					a = &t
				} else if gl.JSONLoad != nil {
					t := *gl.JSONLoad
					t.Input = f
					a = &t
				} else if gl.TableLoad != nil {
					t := *gl.TableLoad
					t.Input = f
					a = &t
				}
				o, err := a.Start(task)
				if err == nil {
					for i := range o {
						if gl.StoreFilename != "" {
							i[gl.StoreFilename] = filepath.Base(f)
						}
						out <- i
					}
				}
			}
		}()
		return out, nil
	}
	return nil, fmt.Errorf("Not found")
}

func (gl *GlobLoadStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(gl.Input) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

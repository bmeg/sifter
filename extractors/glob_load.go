package extractors

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type GlobLoadStep struct {
	StoreFilename string         `json:"storeFilename"`
	StoreFilepath string         `json:"storeFilepath"`
	Input         string         `json:"input" jsonschema_description:"Path of avro object file to transform"`
	Parallelize   bool           `json:"parallelize"`
	XMLLoad       *XMLLoadStep   `json:"xmlLoad"`
	TableLoad     *TableLoadStep `json:"tableLoad" jsonschema_description:"Run transform pipeline on a TSV or CSV"`
	JSONLoad      *JSONLoadStep  `json:"jsonLoad" jsonschema_description:"Run a transform pipeline on a multi line json file"`
	AvroLoad      *AvroLoadStep  `json:"avroLoad" jsonschema_description:"Load data from avro file"`
}

type fileSource struct {
	file   string
	source Source
}

func (gl *GlobLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	input, err := evaluate.ExpressionString(gl.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	logger.Debug("Starting Glob Load: %s", input)
	if gl.XMLLoad != nil || gl.JSONLoad != nil || gl.TableLoad != nil {
		flist, err := filepath.Glob(input)
		if err != nil {
			return nil, err
		}
		pCount := 1
		if gl.Parallelize {
			pCount = 4
		}
		out := make(chan map[string]any, 10*pCount)
		sources := make(chan fileSource, 4*pCount)
		go func() {
			defer close(sources)
			for count, f := range flist {
				logger.Debug("Glob %d of %d", count, len(flist))
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
				sources <- fileSource{source: a, file: f}
			}
		}()
		wg := &sync.WaitGroup{}
		for c := 0; c < pCount; c++ {
			wg.Add(1)
			go func() {
				for a := range sources {
					o, err := a.source.Start(task)
					if err == nil {
						for i := range o {
							if gl.StoreFilename != "" {
								i[gl.StoreFilename] = filepath.Base(a.file)
							}
							if gl.StoreFilepath != "" {
								i[gl.StoreFilepath] = a.file
							}
							out <- i
						}
					}
				}
				wg.Done()
			}()
		}
		go func() {
			wg.Wait()
			close(out)
		}()
		return out, nil
	}
	return nil, fmt.Errorf("not found")
}

func (gl *GlobLoadStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(gl.Input) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

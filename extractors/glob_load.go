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
	Path          string         `json:"path" jsonschema_description:"Path of avro object file to transform"`
	Parallelize   bool           `json:"parallelize"`
	XMLLoad       *XMLLoadStep   `json:"xml"`
	TableLoad     *TableLoadStep `json:"table" jsonschema_description:"Run transform pipeline on a TSV or CSV"`
	JSONLoad      *JSONLoadStep  `json:"json" jsonschema_description:"Run a transform pipeline on a multi line json file"`
	AvroLoad      *AvroLoadStep  `json:"avro" jsonschema_description:"Load data from avro file"`
}

type fileSource struct {
	file   string
	source Source
}

func (gl *GlobLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	input, err := evaluate.ExpressionString(gl.Path, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	logger.Debug("Starting Glob Load", "input", input)
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
				logger.Debug("Glob", "file", f, "num", count, "count", len(flist))
				//var a func()
				var a Source
				if gl.XMLLoad != nil {
					t := *gl.XMLLoad
					t.Path = f
					a = &t
				} else if gl.JSONLoad != nil {
					t := *gl.JSONLoad
					t.Path = f
					a = &t
				} else if gl.TableLoad != nil {
					t := *gl.TableLoad
					t.Path = f
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

func (gl *GlobLoadStep) GetRequiredParams() []config.ParamRequest {
	out := []config.ParamRequest{}
	for _, s := range evaluate.ExpressionIDs(gl.Path) {
		out = append(out, config.ParamRequest{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

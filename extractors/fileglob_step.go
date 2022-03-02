package extractors

import (
	"log"
	"path/filepath"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type FileGlobStep struct {
	Files     []string `json:"files" jsonschema_description:"Array of files (with wildcards) to scan for"`
	Dir       string   `json:"dir"`
	Limit     int      `json:"limit" jsonschema_description:""`
	InputName string   `json:"inputName" jsonschema_description:"variable name the file will be stored in when calling the extraction steps"`
}

func (fs *FileGlobStep) Run(task *task.Task) error {

	log.Printf("FileGlob")
	for _, input := range fs.Files {
		input, err := evaluate.ExpressionString(input, task.Inputs, nil)
		if err != nil {
			return err
		}
		var globPath string
		if fs.Dir == "" {
			globPath, err = task.AbsPath(input)
		} else {
			dir, err := evaluate.ExpressionString(fs.Dir, task.Inputs, nil)
			if err == nil {
				globPath = filepath.Join(dir, input)
			}
		}
		log.Printf("Finding: %s", globPath)
		paths, _ := filepath.Glob(globPath)
		for _, path := range paths {
			log.Printf("Globbed File: %s", path)
			newInputs := map[string]interface{}{}
			for k, v := range task.Inputs {
				newInputs[k] = v
			}
			newInputs[fs.InputName] = path
			/*
				for _, s := range fs.Steps {
					s.Run(task.Runtime, task.SourcePath, newInputs)
				}
			*/
		}
	}
	return nil
}

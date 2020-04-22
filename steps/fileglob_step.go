package steps

import (
  "log"
  "path/filepath"
  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/pipeline"
)
type FileGlobStep struct {
	Files     []string `json:"files"`
	Limit     int      `json:"limit"`
	InputName string   `json:"inputName"`
	Steps     []Step   `json:"steps"`
}

func (fs *FileGlobStep) Run(task *pipeline.Task) error {

  log.Printf("FileGlob")
	for _, input := range fs.Files {
		input, err := evaluate.ExpressionString(input, task.Inputs, nil)
		if err != nil {
			return err
		}
    globPath, err := task.Path(input)
    log.Printf("Finding: %s", globPath)
    paths, _ := filepath.Glob(globPath)
    for _, path := range paths {
      log.Printf("Globbed File: %s", path)
      newInputs := map[string]interface{}{}
      for k,v := range task.Inputs {
        newInputs[k] = v
      }
      newInputs[fs.InputName] = path
      for _, s := range fs.Steps {
        s.Run(task.Runtime, newInputs)
      }
    }
	}
  return nil
}

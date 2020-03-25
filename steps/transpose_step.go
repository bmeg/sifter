
package steps


import (
  "os"
  "io"
  "log"
  "encoding/csv"
  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/pipeline"
)


type TransposeFileStep struct {
  Input   string `json:"input"`
  Output  string `json:"output"`
}


func (ml *TransposeFileStep) Run(task *pipeline.Task) error {

	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
  output, err := evaluate.ExpressionString(ml.Output, task.Inputs, nil)

	inputPath, err := task.Path(input)
  outputPath, err := task.Path(output)

  matrix := [][]string{}

  hd, err := os.Open(inputPath)
  if err != nil {
    return err
  }
  defer hd.Close()

  r := csv.NewReader(hd)
  r.Comma = '\t'

  for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
      log.Printf("Error %s", err)
			break
		}
    matrix = append(matrix, record)
	}

  log.Printf("Writing %s", outputPath)

  ohd, err := os.Create(outputPath)
  w := csv.NewWriter(ohd)
  w.Comma = '\t'

  l := len(matrix[0])
  h := len(matrix)
  for i := 0; i < l; i++ {
    o := make([]string, h)
    for j := 0; j < h; j++ {
      o[j] = matrix[j][i]
    }
    w.Write(o)
  }
  w.Flush()
  return nil
}

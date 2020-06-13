
package steps


import (
  "os"
  "io"
  "log"
  "strings"

  "compress/gzip"

  "encoding/csv"
  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/pipeline"
)


type TransposeFileStep struct {
  Input   string  `json:"input" jsonschema_description:"TSV to transpose"`
  Output  string  `json:"output" jsonschema_description:"Where transpose output should be stored"`
  LineSkip int    `json:"lineSkip" jsonschema_description:"Number of header lines to skip"`
}


func (ml *TransposeFileStep) Run(task *pipeline.Task) error {

	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
  output, err := evaluate.ExpressionString(ml.Output, task.Inputs, nil)

	inputPath, err := task.Path(input)
  outputPath, err := task.Path(output)

  matrix := [][]string{}

  fhd, err := os.Open(inputPath)
  if err != nil {
    return err
  }
  defer fhd.Close()

  var hd io.Reader
  if strings.HasSuffix(input, ".gz") {
    hd, err = gzip.NewReader(fhd)
    if err != nil {
      return err
    }
  } else {
    hd = fhd
  }

  lineSkip := ml.LineSkip

  r := csv.NewReader(hd)
  r.Comma = '\t'
  r.FieldsPerRecord = -1

  for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
    if lineSkip > 0 {
      lineSkip--
    } else {
  		if err != nil {
        log.Printf("Error %s", err)
  			break
  		}
      matrix = append(matrix, record)
    }
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


package steps


import (
  "os"
  "io"
  "log"
  "fmt"
  "strings"
  "sync"

  "compress/gzip"

  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/golib"
  "github.com/bmeg/sifter/transform"
  "github.com/bmeg/sifter/pipeline"
)


type TableLoadStep struct {
  Input         string                  `json:"input"`
	RowSkip       int                     `json:"rowSkip"`
  SkipIfMissing bool                    `json:"skipIfMissing"`
  Columns       []string                `json:"columns"`
  Transform     transform.TransformPipe `json:"transform"`
  Sep           string                  `json:"sep"`
}

func (ml *TableLoadStep) Run(task *pipeline.Task) error {
  log.Printf("Starting Table Load")
	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		if ml.SkipIfMissing {
			return nil
		}
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)
	fhd, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer fhd.Close()

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return err
		}
	} else {
    hd = fhd
  }


  r := golib.CSVReader{}
  if ml.Sep == "" {
    r.Comma = "\t"
  } else {
    r.Comma = ml.Sep
  }
  r.Comment = "#"

  var columns []string
  if ml.Columns != nil {
    columns = ml.Columns
  }

  procChan := make(chan map[string]interface{}, 100)
  wg := &sync.WaitGroup{}

  if err := ml.Transform.Init(task); err != nil {
    return err
  }

  out, err := ml.Transform.Start( procChan, task, wg )
  if err != nil {
    return err
  }
  go func() {
    for range out {}
  }()

  rowSkip := ml.RowSkip

  inputStream, err := golib.ReadLines(hd)
  if err != nil {
    log.Printf("Error %s", err)
    return err
  }

  for record := range r.Read(inputStream) {
    if rowSkip > 0 {
      rowSkip--
    } else {
      if columns == nil {
        columns = record
      } else {
        o := map[string]interface{}{}
        if len(record) >= len(columns) {
          for i, n := range columns {
            o[n] = record[i]
          }
          procChan <- o
        }
      }
    }
	}

  log.Printf("Done Loading")
  close(procChan)
  wg.Wait()
  ml.Transform.Close()

	return nil
}

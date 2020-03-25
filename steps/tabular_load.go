
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

  procChan := []chan map[string]interface{}{}
  wg := &sync.WaitGroup{}
  for _, trans := range ml.Transform {
    i := make(chan map[string]interface{}, 100)
    trans.Start(i, task, wg)
    procChan = append(procChan, i)
  }
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
          for _, c := range procChan {
            c <- o
          }
        }
      }
    }
	}

  log.Printf("Done Loading")
  for _, c := range procChan {
    close(c)
  }
  wg.Wait()

	return nil
}

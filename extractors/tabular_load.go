package extractors

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"compress/gzip"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/readers"
	"github.com/bmeg/sifter/task"
)

type TableLoadStep struct {
	Input        string   `json:"input" jsonschema_description:"TSV to be transformed"`
	RowSkip      int      `json:"rowSkip" jsonschema_description:"Number of header rows to skip"`
	Columns      []string `json:"columns" jsonschema_description:"Manually set names of columns"`
	ExtraColumns string   `json:"extraColumns" jsonschema_description:"Columns beyond originally declared columns will be placed in this array"`
	Sep          string   `json:"sep" jsonschema_description:"Separator \\t for TSVs or , for CSVs"`
}

func (ml *TableLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	log.Printf("Starting Table Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetInputs(), nil)
	inputPath, err := task.AbsPath(input)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading table: %s", inputPath)
	fhd, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer fhd.Close()

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return nil, err
		}
	} else {
		hd = fhd
	}

	r := readers.CSVReader{}
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

	procChan := make(chan map[string]interface{}, 25)

	rowSkip := ml.RowSkip

	inputStream, err := readers.ReadLines(hd)
	if err != nil {
		log.Printf("Error %s", err)
		return nil, err
	}

	go func() {
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
						if ml.ExtraColumns != "" {
							if len(record) > len(columns) {
								o[ml.ExtraColumns] = record[len(columns):]
							}
						}
						procChan <- o
					}
				}
			}
		}

		log.Printf("Done Loading")
		close(procChan)
	}()

	return procChan, nil
}

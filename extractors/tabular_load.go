package extractors

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"compress/gzip"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type TableLoadStep struct {
	Input        string   `json:"input" jsonschema_description:"TSV to be transformed"`
	RowSkip      int      `json:"rowSkip" jsonschema_description:"Number of header rows to skip"`
	Columns      []string `json:"columns" jsonschema_description:"Manually set names of columns"`
	ExtraColumns string   `json:"extraColumns" jsonschema_description:"Columns beyond originally declared columns will be placed in this array"`
	Sep          string   `json:"sep" jsonschema_description:"Separator \\t for TSVs or , for CSVs"`
	Comment      string   `json:"comment"`
	LazyQuotes   bool     `json:"lazyQuotes"`
}

func arrayContains(a []string, b string) bool {
	for _, c := range a {
		if c == b {
			return true
		}
	}
	return false
}

func buildUniqueArray(src []string) []string {
	out := []string{}
	for _, n := range src {
		name := n
		attempt := 1
		for arrayContains(out, name) {
			name = fmt.Sprintf("%s_%d", n, attempt)
			attempt++
		}
		out = append(out, name)
	}
	return out
}

func (ml *TableLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	log.Printf("Starting Table Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	inputPath, _ := task.AbsPath(input)

	if s, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", inputPath)
	} else if s.IsDir() {
		return nil, fmt.Errorf("input not a file: %s", inputPath)
	}
	log.Printf("Loading table: %s", inputPath)

	var inputStream io.ReadCloser
	if gfile, err := os.Open(inputPath); err == nil {
		if strings.HasSuffix(inputPath, ".gz") {
			inp, err := gzip.NewReader(gfile)
			if err != nil {
				return nil, err
			}
			inputStream = inp
		} else {
			inputStream = gfile
		}
	}
	if err != nil {
		return nil, err
	}

	if ml.Sep == "" {
		ml.Sep = "\t"
	}

	tsvReader := csv.NewReader(inputStream)
	tsvReader.Comma = rune(ml.Sep[0])
	tsvReader.LazyQuotes = ml.LazyQuotes
	tsvReader.Comment = '#'
	if ml.Comment != "" {
		tsvReader.Comment = []rune(ml.Comment)[0]
	}
	var columns []string
	if ml.Columns != nil {
		columns = ml.Columns
		tsvReader.FieldsPerRecord = len(ml.Columns)
	}

	procChan := make(chan map[string]interface{}, 25)

	rowSkip := ml.RowSkip

	go func() {
		defer inputStream.Close()
		//log.Printf("STARTING READ: %#v", inputStream)
		for {
			record, err := tsvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				if pe, ok := err.(*csv.ParseError); ok {
					if pe.Err == csv.ErrFieldCount {
						//log.Printf("Field count %d != %d", len(record), len(columns))
						if len(record) < len(columns) {
							c := len(columns) - len(record)
							for i := 0; i < c; i++ {
								record = append(record, "")
							}
							err = nil
						}
					} else if pe.Err == csv.ErrQuote {
						log.Printf("quote error: %s", record)
					} else if pe.Err == csv.ErrBareQuote {
						log.Printf("bare quote error: %s", record)
					}
				}
			}

			if err == nil {
				if rowSkip > 0 {
					rowSkip--
				} else {
					if columns == nil {
						columns = buildUniqueArray(record)
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
			} else {
				log.Printf("Error: %s", err)
			}
		}
		log.Printf("Done Loading")
		close(procChan)
	}()

	return procChan, nil
}

func (ml *TableLoadStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(ml.Input) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

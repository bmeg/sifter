package extractors

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	"github.com/xwb1989/sqlparser"
)

type SQLDumpStep struct {
	Input  string   `json:"input" jsonschema_description:"Path to the SQL dump file"`
	Tables []string `json:"tables" jsonschema_description:"Array of transforms for the different tables in the SQL dump"`
}

func (ml *SQLDumpStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {

	input, err := evaluate.ExpressionString(ml.Input, task.GetInputs(), nil)
	inputPath, err := task.AbsPath(input)
	log.Printf("Starting SQLDump Load: %s", inputPath)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)
	fhd, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return nil, err
		}
	} else {
		hd = fhd
	}

	out := make(chan map[string]any, 100)
	tables := map[string]bool{}
	for t := range ml.Tables {
		tables[ml.Tables[t]] = true
	}

	go func() {
		defer fhd.Close()
		defer close(out)
		tableColumns := map[string][]string{}
		tokens := sqlparser.NewTokenizer(hd)
		for {
			stmt, err := sqlparser.ParseNext(tokens)
			if err == io.EOF {
				break
			}
			switch stmt := stmt.(type) {
			case *sqlparser.DDL:
				if stmt.Action == "create" {
					fmt.Printf("SQL Parser found: Table Create: %s\n", stmt.NewName.Name.CompliantName())
					columns := []string{}
					for _, col := range stmt.TableSpec.Columns {
						name := col.Name.CompliantName()
						columns = append(columns, name)
					}
					fmt.Printf("%s\n", columns)
					tableColumns[stmt.NewName.Name.CompliantName()] = columns
				}
			case *sqlparser.Insert:
				//fmt.Printf("Inserting into: %s\n", stmt.Table.Name)

				tableName := stmt.Table.Name.CompliantName()

				if _, ok := tables[tableName]; ok {
					cols := tableColumns[tableName]
					if irows, ok := stmt.Rows.(sqlparser.Values); ok {
						for _, row := range irows {
							data := map[string]interface{}{}
							for i := range row {
								if sval, ok := row[i].(*sqlparser.SQLVal); ok {
									data[cols[i]] = string(sval.Val)
								}
							}
							out <- map[string]any{"table": tableName, "data": data}
						}
					}
				} else {
					log.Printf("WARNING: Other sql.InsertValue: %s", tableName)
				}
			}
		}
	}()

	return out, nil
}

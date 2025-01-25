package extractors

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	"vitess.io/vitess/go/vt/sqlparser"
)

type SQLDumpStep struct {
	Input  string   `json:"input" jsonschema_description:"Path to the SQL dump file"`
	Tables []string `json:"tables" jsonschema_description:"Array of transforms for the different tables in the SQL dump"`
}

func (ml *SQLDumpStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {

	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	inputPath, _ := task.AbsPath(input)
	log.Printf("Starting SQLDump Load: %s", inputPath)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", input)
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
	if ml.Tables != nil {
		for t := range ml.Tables {
			tables[ml.Tables[t]] = true
		}
	}

	go func() {
		defer fhd.Close()
		defer close(out)
		tableColumns := map[string][]string{}
		data, _ := io.ReadAll(hd)
		parser := sqlparser.Parser{}
		tokens := parser.NewStringTokenizer(string(data))
		for {
			stmt, err := sqlparser.ParseNext(tokens)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("read error: %s", err)
			}
			switch stmt := stmt.(type) {
			case *sqlparser.CreateTable:
				fmt.Printf("SQL Parser found: Table Create: %s\n", stmt.Table.Name.CompliantName())
				fmt.Printf("%#v\n", stmt)
				columns := []string{}
				if stmt.TableSpec != nil {
					for _, col := range stmt.TableSpec.Columns {
						name := col.Name.CompliantName()
						columns = append(columns, name)
					}
				}
				fmt.Printf("%s\n", columns)
				tableColumns[stmt.Table.Name.CompliantName()] = columns

			case *sqlparser.Insert:
				//fmt.Printf("Inserting into: %s\n", stmt.Table.Name)

				t, _ := stmt.Table.TableName()
				tableName := t.Name.CompliantName()

				if _, ok := tables[tableName]; ok || len(tables) == 0 {
					cols := tableColumns[tableName]
					if irows, ok := stmt.Rows.(sqlparser.Values); ok {
						for _, row := range irows {
							data := map[string]interface{}{}
							for i := range row {
								if sval, ok := row[i].(*sqlparser.Literal); ok {
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

func (ml *SQLDumpStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(ml.Input) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

package extractors

import (
  "os"
  "io"
  "fmt"
  "log"
  "strings"
  "compress/gzip"
  "github.com/xwb1989/sqlparser"
  "github.com/bmeg/sifter/transform"
  "github.com/bmeg/sifter/pipeline"
  "github.com/bmeg/sifter/evaluate"
)



type TableTransform struct {
  Table         string                  `json:"table" jsonschema_description:"Name of the SQL file to transform"`
  Transform     transform.TransformPipe `json:"transform" jsonschema_description:"The transform pipeline"`
}


type SQLDumpStep struct {
  Input         string                  `json:"input" jsonschema_description:"Path to the SQL dump file"`
  Tables        []TableTransform        `json:"tables" jsonschema_description:"Array of transforms for the different tables in the SQL dump"`
  SkipIfMissing bool                    `json:"skipIfMissing" jsonschema_description:"Option to skip without fail if input file does not exist"`
}


func (ml *SQLDumpStep) Run(task *pipeline.Task) error {

  log.Printf("Starting SQLDump Load")
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
          fmt.Printf("Table Create: %s\n", stmt.NewName.Name. CompliantName())
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
        data := map[string]interface{}{}

        cols := tableColumns[stmt.Table.Name.CompliantName()]
        if irows, ok := stmt.Rows.(sqlparser.Values); ok {
          for _, row := range irows {
            for i := range row {
              if sval, ok := row[i].(*sqlparser.SQLVal); ok {
                data[cols[i]] = string(sval.Val)
              }
            }
          }
          fmt.Printf("%s - %s\n", stmt.Table.Name, data)
        }
    }
  }
  return nil
}

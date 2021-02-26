package extractors

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/pipeline"
	"github.com/bmeg/sifter/transform"
	"github.com/xwb1989/sqlparser"
)

type TableTransform struct {
	Name      string         `json:"name" jsonschema_description:"Name of the SQL file to transform"`
	Transform transform.Pipe `json:"transform" jsonschema_description:"The transform pipeline"`
}

type QueryTransform struct {
	Query     string         `json:"query" jsonschema_description:"SQL select query to use as input"`
	Transform transform.Pipe `json:"transform" jsonschema_description:"The transform pipeline"`
}

type SQLDumpStep struct {
	Input         string           `json:"input" jsonschema_description:"Path to the SQL dump file"`
	Tables        []TableTransform `json:"tables" jsonschema_description:"Array of transforms for the different tables in the SQL dump"`
	SkipIfMissing bool             `json:"skipIfMissing" jsonschema_description:"Option to skip without fail if input file does not exist"`
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

	wg := &sync.WaitGroup{}
	chanMap := map[string][]chan map[string]interface{}{}

	for t := range ml.Tables {
		procChan := make(chan map[string]interface{}, 100)
		if err := ml.Tables[t].Transform.Init(task); err != nil {
			return err
		}
		out, err := ml.Tables[t].Transform.Start(procChan, task, wg)
		if err != nil {
			return err
		}
		go func() {
			for range out {
			}
		}()
		log.Printf("Adding transform pipe for table: %s", ml.Tables[t].Name)
		if x, ok := chanMap[ml.Tables[t].Name]; ok {
			chanMap[ml.Tables[t].Name] = append(x, procChan)
		} else {
			chanMap[ml.Tables[t].Name] = []chan map[string]interface{}{procChan}
		}
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

			cols := tableColumns[stmt.Table.Name.CompliantName()]
			if irows, ok := stmt.Rows.(sqlparser.Values); ok {
				for _, row := range irows {
					data := map[string]interface{}{}
					for i := range row {
						if sval, ok := row[i].(*sqlparser.SQLVal); ok {
							data[cols[i]] = string(sval.Val)
						}
					}
					if x, ok := chanMap[stmt.Table.Name.CompliantName()]; ok {
						//fmt.Printf("%s - %s\n", stmt.Table.Name.CompliantName(), data)
						for i := range x {
							x[i] <- data
						}
					} else {
						//log.Printf("Skip: %s", stmt.Table.Name.CompliantName())
					}
				}
			} else {
				log.Printf("WARNING: Other sql.InsertValue")
			}
		}
	}
	for i := range chanMap {
		for j := range chanMap[i] {
			close(chanMap[i][j])
		}
	}
	wg.Wait()
	for t := range ml.Tables {
		ml.Tables[t].Transform.Close()
	}

	return nil
}

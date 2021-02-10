package extractors

import (
  "os"
  "log"
	"database/sql"
	"fmt"
  "sync"

	_ "github.com/mattn/go-sqlite3"
  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/pipeline"
)

type SQLiteStep struct {
	Input         string           `json:"input" jsonschema_description:"Path to the SQLite file"`
	Tables        []TableTransform `json:"tables" jsonschema_description:"Array of transforms for the different tables in the SQLite"`
	SkipIfMissing bool             `json:"skipIfMissing" jsonschema_description:"Option to skip without fail if input file does not exist"`
}

func (ml *SQLiteStep) Run(task *pipeline.Task) error {

	log.Printf("Starting SQLite Load")
	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		if ml.SkipIfMissing {
			return nil
		}
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("SQLite Loading: %s", inputPath)

	db, err := sql.Open("sqlite3", inputPath)
	if err != nil {
		return err
	}

	for t := range ml.Tables {
		rows, err := db.Query("select * from " + ml.Tables[t].Name)
		if err == nil {
      log.Printf("Scanning table %s", ml.Tables[t].Name)
			wg := &sync.WaitGroup{}
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
      colNames, err := rows.Columns()
      readCols := make([]interface{}, len(colNames))
      writeCols := make([]string, len(colNames))
      for i, _ := range writeCols {
        readCols[i] = &writeCols[i]
      }
			for rows.Next() {
				err := rows.Scan(readCols...)
        if err != nil {
          log.Printf("scan error: %s", err)
        } else {
          o := map[string]interface{}{}
          for i := range colNames {
            o[colNames[i]] = writeCols[i]
          }
          fmt.Printf("%#v\n", o)
        }
			}
			ml.Tables[t].Transform.Close()
			wg.Wait()
		} else {
      log.Printf("SQLite table read error: %s", err)
    }
	}
	return nil
}

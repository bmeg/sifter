package extractors

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	"github.com/bmeg/sifter/transform"
	_ "github.com/mattn/go-sqlite3" // Adding sqlite3 to the SQL driver list
)

type QueryTransform struct {
	Query     string         `json:"query" jsonschema_description:"SQL select query to use as input"`
	Transform transform.Pipe `json:"transform" jsonschema_description:"The transform pipeline"`
}

type SQLiteStep struct {
	Input         string           `json:"input" jsonschema_description:"Path to the SQLite file"`
	Tables        []TableTransform `json:"tables" jsonschema_description:"Array of transforms for the different tables in the SQLite"`
	Queries       []QueryTransform `json:"queries" jsonschema_description:"SQL select statement based input"`
	SkipIfMissing bool             `json:"skipIfMissing" jsonschema_description:"Option to skip without fail if input file does not exist"`
}

func processQuery(rows *sql.Rows, trans transform.Pipe, task task.RuntimeTask) error {
	wg := &sync.WaitGroup{}
	procChan := make(chan map[string]interface{}, 100)
	if err := trans.Init(task); err != nil {
		return err
	}
	out, err := trans.Start(procChan, task, wg)
	if err != nil {
		return err
	}
	go func() {
		for range out {
		}
	}()
	colNames, err := rows.Columns()
	readCols := make([]any, len(colNames))
	writeCols := make([]sql.NullString, len(colNames))
	for i := range writeCols {
		readCols[i] = &writeCols[i]
	}
	for rows.Next() {
		err := rows.Scan(readCols...)
		if err != nil {
			log.Printf("scan error: %s", err)
		} else {
			o := map[string]interface{}{}
			for i := range colNames {
				if writeCols[i].Valid {
					o[colNames[i]] = writeCols[i].String
				}
			}
			procChan <- o
		}
	}
	close(procChan)
	trans.Close()
	wg.Wait()
	return nil
}

func (ml *SQLiteStep) Start(task.RuntimeTask) (chan map[string]interface{}, error) {

	log.Printf("Starting SQLite Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetInputs(), nil)
	inputPath, err := task.AbsPath(input)

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
			processQuery(rows, ml.Tables[t].Transform, task)
		} else {
			log.Printf("SQLite table read error: %s", err)
		}
	}

	for t := range ml.Queries {
		rows, err := db.Query(ml.Queries[t].Query)
		if err == nil {
			log.Printf("Scanning table %s", ml.Queries[t].Query)
			processQuery(rows, ml.Queries[t].Transform, task)
		} else {
			log.Printf("SQLite table read error: %s", err)
		}
	}

	return nil
}

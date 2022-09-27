package extractors

import (
	"database/sql"
	"log"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	_ "github.com/mattn/go-sqlite3" // Adding sqlite3 to the SQL driver list
)

type SQLiteStep struct {
	Input string `json:"input" jsonschema_description:"Path to the SQLite file"`
	Query string `json:"query" jsonschema_description:"SQL select statement based input"`
}

func (ml *SQLiteStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	inputPath, err := task.AbsPath(input)

	log.Printf("SQLite Loading: %s", inputPath)

	db, err := sql.Open("sqlite3", inputPath)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ml.Query)
	if err != nil {
		log.Printf("SQLite table read error: %s", err)
		return nil, err
	}

	procChan := make(chan map[string]interface{}, 100)
	go func() {
		colNames, _ := rows.Columns()
		readCols := make([]interface{}, len(colNames))
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
	}()

	return procChan, nil
}

func (ml *SQLiteStep) GetConfigFields() []config.ConfigVar {
	out := []config.ConfigVar{}
	for _, s := range evaluate.ExpressionIDs(ml.Input) {
		out = append(out, config.ConfigVar{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

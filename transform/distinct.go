package transform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	badger "github.com/dgraph-io/badger/v2"
)

type DistinctStep struct {
	Value string `json:"value"`
}

type distinctProcess struct {
	config DistinctStep
	task   task.RuntimeTask
	db     *badger.DB
	dir    string
}

func (ds DistinctStep) Init(task task.RuntimeTask) (Processor, error) {
	log.Printf("Starting Distinct: %s", ds.Value)
	tdir := task.TempDir()
	opts := badger.DefaultOptions(filepath.Join(tdir, "badger"))
	opts.ValueDir = filepath.Join(tdir, "badger")
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &distinctProcess{ds, task, db, tdir}, nil
}

func (ds *distinctProcess) Process(i map[string]any) []map[string]any {
	out := []map[string]any{}

	keyStr, err := evaluate.ExpressionString(ds.config.Value, ds.task.GetConfig(), i)
	if err == nil {
		ds.db.Update(func(txn *badger.Txn) error {
			key := []byte(keyStr)
			_, err := txn.Get(key)
			if err == badger.ErrKeyNotFound {
				fmt.Printf("Distinct: %s\n", keyStr)
				out = append(out, i)
				txn.Set(key, []byte{})
			}
			return nil
		})
	} else {
		log.Printf("Distinct field error %s", err)
	}
	return out
}

func (ds *distinctProcess) Close() {
	ds.db.Close()
	log.Printf("Closing DB")
	if err := os.RemoveAll(ds.dir); err != nil {
		log.Printf("%s", err)
	}
}

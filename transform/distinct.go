package transform

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	badger "github.com/dgraph-io/badger/v2"
)

type DistinctStep struct {
	Field string `json:"field"`
	Steps Pipe   `json:"steps"`
	db    *badger.DB
}

func (ds *DistinctStep) Init(task task.RuntimeTask) {
	log.Printf("Starting Distinct: %s", ds.Field)
	tdir := task.TempDir()
	opts := badger.DefaultOptions(filepath.Join(tdir, "badger"))
	opts.ValueDir = filepath.Join(tdir, "badger")
	var err error
	ds.db, err = badger.Open(opts)
	if err != nil {
		log.Printf("%s", err)
	}
}

func (ds *DistinctStep) Start(in chan map[string]interface{}, task task.RuntimeTask, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	out := make(chan map[string]interface{}, 10)

	go func() {
		defer close(out)
		ds.db.Update(func(txn *badger.Txn) error {
			for i := range in {
				keyStr, err := evaluate.ExpressionString(ds.Field, task.GetInputs(), i)
				if err == nil {
					key := []byte(keyStr)
					_, err := txn.Get(key)
					if err == badger.ErrKeyNotFound {
						out <- i
						txn.Set(key, []byte{})
					}
				} else {
					log.Printf("Distinct field error %s", err)
				}
			}
			return nil
		})
	}()
	return out, nil
}

func (ds *DistinctStep) Close() {
	ds.db.Close()
}

package transform


import (
  "log"
  "sync"
  "path/filepath"
  "github.com/dgraph-io/badger/v2"
  "github.com/bmeg/sifter/pipeline"
  "github.com/bmeg/sifter/evaluate"
)


type DistinctStep struct {
  Field  string `json:"field"`
  Steps  Pipe   `json:"steps"`
  db     *badger.DB
}

func (ds *DistinctStep) Init(task *pipeline.Task) {
	log.Printf("Starting Distinct: %s", ds.Field)
	tdir := task.TempDir()
	opts := badger.DefaultOptions(filepath.Join(tdir, "badger"))
	opts.ValueDir = filepath.Join(tdir, "badger")
  var err error
	ds.db, err = badger.Open(opts)
	if err != nil {
		log.Printf("%s", err)
	}
  ds.Steps.Init(task)
}

func (ds *DistinctStep) Start(in chan map[string]interface{}, task *pipeline.Task, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	out := make(chan map[string]interface{}, 10)

  inChan := make(chan map[string]interface{}, 100)
	tout, _ := ds.Steps.Start(inChan, task.Child("distinct"), wg)
	go func() {
		//Distinct does not emit the output of its sub pipeline, but it has to digest it
		for range tout { }
	}()


	go func() {
		defer close(out)
    defer close(inChan)

    ds.db.Update( func(txn *badger.Txn) error {
      for i := range in {
        out <- i
        keyStr, err := evaluate.ExpressionString(ds.Field, task.Inputs, i)
        if err == nil {
          key := []byte(keyStr)
          _, err := txn.Get(key)
          if err == badger.ErrKeyNotFound {
	           inChan <- i
             txn.Set(key, []byte{})
          }
        } else {
          log.Printf("Distinct field error %s", err )
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

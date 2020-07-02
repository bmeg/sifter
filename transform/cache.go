package transform

import (
  "log"
  "fmt"
  "sync"
  "github.com/bmeg/sifter/pipeline"

  "github.com/cnf/structhash"
)

func (cs *CacheStep) Init(task *pipeline.Task) error {
	return cs.Transform.Init(task)
}

func (cs *CacheStep) Start(in chan map[string]interface{}, task *pipeline.Task, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	log.Printf("Starting Cache: %s", task.Name)

	ds, err := task.GetDataStore()
	if err != nil {
		log.Printf("Cache setup error: %s", err)
	}

	if ds == nil {
		log.Printf("No cache setup")
		out, err := cs.Transform.Start(in, task, wg)
		return out, err
	}

	out := make(chan map[string]interface{}, 10)
	go func() {
		defer close(out)
		for i := range in {
			hash, err := structhash.Hash(i, 1)
			if err == nil {
				key := fmt.Sprintf("%s.%s", task.Name, hash)
				log.Printf("Cache Key: %s.%s", task.Name, hash)
				if ds.HasRecordStream(key) {
					log.Printf("Cache Hit")
					for j := range ds.GetRecordStream(key) {
						out <- j
					}
				} else {
					log.Printf("Cache Miss")

					manIn := make(chan map[string]interface{}, 10)
					manIn <- i
					close(manIn)

					cacheIn := make(chan map[string]interface{}, 10)
					go ds.SetRecordStream(key, cacheIn)

					newWG := &sync.WaitGroup{}

					tOut, _ := cs.Transform.Start(manIn, task, newWG)
					for j := range tOut {
						log.Printf("Cache Calc out: %s", j)
						cacheIn <- j
						out <- j
					}
					close(cacheIn)
				}
			} else {
				log.Printf("Hashing Error")
			}
		}
	}()

	return out, nil
}

func (cs *CacheStep) Close() {

}

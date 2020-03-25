package transform


import (
  "os"
  "encoding/json"
  "encoding/binary"
  "path/filepath"
  "bytes"
  "sync"
  "log"
  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/pipeline"
  "github.com/dgraph-io/badger"
)

type MapStep struct {
  Method string `json:"method"`
  Python string `json:"python"`
  pyCode *evaluate.PyCode
}


func (ms *MapStep) Start(task *pipeline.Task, wg *sync.WaitGroup) {
  log.Printf("Starting Map: %s", ms.Python)
  c, err := evaluate.PyCompile(ms.Python)
  if err != nil {
    log.Printf("%s", err)
  }
  ms.pyCode = c
}

func (ms *MapStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
  out := ms.pyCode.Evaluate(ms.Method, i)
  return out
}


type ReduceStep struct {
  Field  string `json:"field"`
  Method string `json:"method"`
  Python string `json:"python"`
  pyCode *evaluate.PyCode
  dump   *os.File
  db     *badger.DB
  batch  *badger.WriteBatch
}


func (ms *ReduceStep) Start(task *pipeline.Task, wg *sync.WaitGroup) {
  log.Printf("Starting Reduce: %s", ms.Python)
  c, err := evaluate.PyCompile(ms.Python)
  if err != nil {
    log.Printf("%s", err)
  }
  ms.pyCode = c
  tdir := task.TempDir()
  tfile := filepath.Join(tdir, "dump.data")
  log.Printf("Reduce file: %s", tfile)
  ms.dump, err = os.Create(tfile)
  if err != nil {
    log.Printf("%s", err)
  }

  opts := badger.DefaultOptions(filepath.Join(tdir, "badger"))
  opts.ValueDir = filepath.Join(tdir, "badger")
  ms.db, err = badger.Open(opts)
  if err != nil {
   log.Printf("%s", err)
  }
  ms.batch = ms.db.NewWriteBatch()
}

func (ms *ReduceStep) Add(i map[string]interface{}, task *pipeline.Task) {
  d, _ := json.Marshal(i)

  dKey, _ := evaluate.ExpressionString(ms.Field, task.Inputs, i)
  dSize := uint64(len(d))
  stat, _ := ms.dump.Stat()
  dPos := uint64(stat.Size())

  bPos := make([]byte, 8)
  binary.BigEndian.PutUint64(bPos, dPos)
  bSize := make([]byte, 8)
  binary.BigEndian.PutUint64(bSize, dSize)

  key := bytes.Join( [][]byte{ []byte(dKey), bPos }, []byte{} )
  ms.batch.Set(key, bSize)
  ms.dump.Write(d)
  ms.dump.Write([]byte("\n"))
}

type jsonData struct {
  key string
  json []byte
  data map[string]interface{}
}

func (ms *ReduceStep) Run() chan map[string]interface{} {
  ms.batch.Flush()
  log.Printf("Starting Reduce")

  jsonChan := make(chan jsonData, 100)
  go func() {
    defer close(jsonChan)
    ms.db.View(func(txn *badger.Txn) error {
      opts := badger.DefaultIteratorOptions
      opts.PrefetchSize = 10
      it := txn.NewIterator(opts)
      defer it.Close()
      for it.Rewind(); it.Valid(); it.Next() {
        item := it.Item()
        k := item.Key()
        var dSize uint64 = 0
        item.Value(func(v []byte) error {
          dSize = binary.BigEndian.Uint64(v)
          return nil
        })
        key := k[0:len(k)-8]
        bPos := k[len(k)-8:len(k)]
        dPos := binary.BigEndian.Uint64(bPos)
        data := make([]byte, dSize)
        ms.dump.ReadAt(data, int64(dPos))
        jsonChan <- jsonData{key:string(key), json:data}
      }
      ms.dump.Close()
      ms.db.Close()
      return nil
    })
  }()

  dataChan := make(chan jsonData, 100)
  go func() {
    defer close(dataChan)
    for b := range jsonChan {
      o := map[string]interface{}{}
      if err := json.Unmarshal(b.json, &o); err == nil {
        dataChan <- jsonData{key:b.key, data:o}
      }
    }
  }()

  out := make(chan map[string]interface{}, 100)
  go func() {
    key := ""
    var last map[string]interface{}
    for d := range dataChan {
      if d.key != key {
        if key != "" {
          out <- last
        }
        key = d.key
        last = d.data
      } else {
        out := ms.pyCode.Evaluate(ms.Method, last, d.data)
        last = out
      }
    }
    if key != "" {
      out <- last
    }
    defer close(out)
  }()
  return out
}

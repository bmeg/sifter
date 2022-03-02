package transform

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"

	//"sync"
	"log"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	badger "github.com/dgraph-io/badger/v2"
)

type MapStep struct {
	Method  string `json:"method" jsonschema_description:"Name of function to call"`
	Python  string `json:"python" jsonschema_description:"Python code to be run"`
	GPython string `json:"gpython" jsonschema_description:"Python code to be run using GPython"`
	proc    evaluate.Processor
}

func (ms *MapStep) Init(task task.RuntimeTask) {
	if ms.Python != "" {
		log.Printf("Init Map: %s", ms.Python)
		e := evaluate.GetEngine("python", task.WorkDir())
		c, err := e.Compile(ms.Python, ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		ms.proc = c
	} else if ms.GPython != "" {
		log.Printf("Init Map: %s", ms.GPython)
		e := evaluate.GetEngine("gpython", task.WorkDir())
		c, err := e.Compile(ms.GPython, ms.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		ms.proc = c
	}
}

func (ms *MapStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {
	out, err := ms.proc.Evaluate(i)
	if err != nil {
		log.Printf("Map Step error: %s", err)
	}
	return out
}

func (ms *MapStep) Close() {
	ms.proc.Close()
}

type ReduceStep struct {
	Field    string                  `json:"field"`
	Method   string                  `json:"method"`
	Python   string                  `json:"python"`
	InitData *map[string]interface{} `json:"init"`
	proc     evaluate.Processor
	dump     *os.File
	db       *badger.DB
	batch    *badger.WriteBatch
}

func (ms *ReduceStep) Init(task task.RuntimeTask) {
	log.Printf("Starting Reduce: %s", ms.Python)
	e := evaluate.GetEngine(DefaultEngine, task.WorkDir())
	c, err := e.Compile(ms.Python, ms.Method)
	if err != nil {
		log.Printf("%s", err)
	}
	ms.proc = c
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

func (ms *ReduceStep) Add(i map[string]interface{}, task task.RuntimeTask) {
	d, _ := json.Marshal(i)

	dKey, _ := evaluate.ExpressionString(ms.Field, task.GetInputs(), i)
	dSize := uint64(len(d))
	stat, _ := ms.dump.Stat()
	dPos := uint64(stat.Size())

	bPos := make([]byte, 8)
	binary.BigEndian.PutUint64(bPos, dPos)
	bSize := make([]byte, 8)
	binary.BigEndian.PutUint64(bSize, dSize)

	key := bytes.Join([][]byte{[]byte(dKey), bPos}, []byte{})
	ms.batch.Set(key, bSize)
	ms.dump.Write(d)
	ms.dump.Write([]byte("\n"))
}

type jsonData struct {
	key  string
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
				key := k[0 : len(k)-8]
				bPos := k[len(k)-8:]
				dPos := binary.BigEndian.Uint64(bPos)
				data := make([]byte, dSize)
				ms.dump.ReadAt(data, int64(dPos))
				jsonChan <- jsonData{key: string(key), json: data}
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
				dataChan <- jsonData{key: b.key, data: o}
			}
		}
	}()

	out := make(chan map[string]interface{}, 100)
	go func() {
		defer close(out)
		key := ""
		var last map[string]interface{}
		for d := range dataChan {
			if d.key != key {
				if key != "" {
					out <- last
				}
				key = d.key
				if ms.InitData != nil {
					out, _ := ms.proc.Evaluate(*ms.InitData, d.data)
					last = out
				} else {
					last = d.data
				}
			} else {
				out, _ := ms.proc.Evaluate(last, d.data)
				last = out
			}
		}
		if key != "" {
			out <- last
		}
		ms.proc.Close()
	}()
	return out
}

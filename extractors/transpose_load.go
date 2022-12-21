package extractors

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"compress/gzip"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	"github.com/cockroachdb/pebble"
	//badger "github.com/dgraph-io/badger/v2"
)

type TransposeLoadStep struct {
	Input   string `json:"input" jsonschema_description:"TSV to be transformed"`
	RowSkip int    `json:"rowSkip" jsonschema_description:"Number of header rows to skip"`
	Sep     string `json:"sep" jsonschema_description:"Separator \\t for TSVs or , for CSVs"`
	OnDisk  bool   `json:"onDisk" jsonschema_description:"Do transpose without caching matrix in memory. Takes longer but works on large files"`
}

/*
func (ml *TransposeLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	log.Printf("Starting Table Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	inputPath, _ := task.AbsPath(input)

	if s, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", inputPath)
	} else if s.IsDir() {
		return nil, fmt.Errorf("input not a file: %s", inputPath)
	}
	log.Printf("Loading table: %s", inputPath)
	fhd, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return nil, err
		}
	} else {
		hd = fhd
	}

	r := readers.CSVReader{}
	if ml.Sep == "" {
		r.Comma = "\t"
	} else {
		r.Comma = ml.Sep
	}
	r.Comment = "#"

	var columns []string

	procChan := make(chan map[string]interface{}, 25)

	rowSkip := ml.RowSkip

	inputStream, err := readers.ReadLines(hd)
	if err != nil {
		log.Printf("Error %s", err)
		return nil, err
	}

	go func() {
		defer fhd.Close()
		log.Printf("STARTING READ: %#v %#v", r, inputStream)
		for record := range r.Read(inputStream) {
			if rowSkip > 0 {
				rowSkip--
			} else {
				if columns == nil {
					columns = record
				} else {
					o := map[string]interface{}{}
					if len(record) >= len(columns) {
						for i, n := range columns {
							o[n] = record[i]
						}
						procChan <- o
					}
				}
			}
		}
		log.Printf("Done Loading")
		close(procChan)
	}()

	return procChan, nil
}
*/

func (ml *TransposeLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}
	inputPath, _ := task.AbsPath(input)
	cr := csvReader{inputPath, ml.RowSkip, ml.Sep}
	out := make(chan map[string]any, 10)

	if !ml.OnDisk {
		go transposeInMem(cr, out)
	} else {
		tdir := task.TempDir()
		go transposeOnDisk(tdir, cr, out)
	}
	return out, nil
}

func (ml *TransposeLoadStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(ml.Input) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

type csvReader struct {
	inputPath string
	lineSkip  int
	sep       string
}

func (c csvReader) open() (*csv.Reader, error) {
	fhd, err := os.Open(c.inputPath)
	if err != nil {
		return nil, err
	}
	var hd io.Reader
	if strings.HasSuffix(c.inputPath, ".gz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return nil, err
		}
	} else {
		hd = fhd
	}
	r := csv.NewReader(hd)
	r.Comma = '\t'
	if c.sep != "" {
		r.Comma = []rune(c.sep)[0]
	}
	r.FieldsPerRecord = -1
	for i := 0; i < c.lineSkip; i++ {
		r.Read()
	}
	return r, nil
}

func transposeInMem(c csvReader, out chan map[string]any) error {
	matrix := [][]string{}

	r, err := c.open()
	if err != nil {
		return nil
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error %s", err)
			break
		}
		matrix = append(matrix, record)
	}
	log.Printf("Writing Transpose")
	l := len(matrix[0])
	h := len(matrix)
	columns := make([]string, h)
	for j := 0; j < h; j++ {
		columns[j] = matrix[j][0]
	}
	for i := 1; i < l; i++ {
		o := map[string]interface{}{}
		for j, n := range columns {
			o[n] = matrix[j][i]
		}
		out <- o
	}
	close(out)
	return nil
}

/*
func transposeOnDisk(workdir string, c csvReader, out chan map[string]any) error {
	opts := badger.DefaultOptions(filepath.Join(workdir, "transpose.db"))
	opts.ValueDir = filepath.Join(workdir, "transpose.db")
	db, err := badger.Open(opts)
	if err != nil {
		log.Printf("%s", err)
	}
	batch := db.NewWriteBatch()
	r, err := c.open()
	if err != nil {
		return nil
	}
	rowCount := uint64(0)
	colCount := uint64(0)
	for row := uint64(0); ; row++ {
		log.Printf("Row: %d", row)
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if row == 0 {
			colCount = uint64(len(record))
		}
		bRow := make([]byte, 8)
		binary.BigEndian.PutUint64(bRow, row)

		for col := uint64(0); col < uint64(len(record)); col++ {
			bCol := make([]byte, 8)
			binary.BigEndian.PutUint64(bCol, col)
			//key := bytes.Join([][]byte{[]byte(bCol), bRow}, []byte{})
			key := bytes.Join([][]byte{bRow, bCol}, []byte{})
			batch.Set(key, []byte(record[col]))
		}
		rowCount = row + 1
	}
	batch.Flush()

	log.Printf("%d %d", colCount, rowCount)

	columns := []string{}
	db.View(func(txn *badger.Txn) error {
		bCol := make([]byte, 8)
		binary.BigEndian.PutUint64(bCol, 0)
		for row := uint64(0); row < rowCount; row++ {
			bRow := make([]byte, 8)
			binary.BigEndian.PutUint64(bRow, row)
			key := bytes.Join([][]byte{bRow, bCol}, []byte{})
			item, err := txn.Get(key)
			if err == nil {
				item.Value(func(val []byte) error {
					columns = append(columns, string(val))
					return nil
				})
			}
		}
		return nil
	})

	for col := uint64(1); col < colCount; col++ {
		log.Printf("Writing Row %d", col)
		bCol := make([]byte, 8)
		binary.BigEndian.PutUint64(bCol, col)
		o := []string{}
		db.View(func(txn *badger.Txn) error {
			for row := uint64(0); row < rowCount; row++ {
				bRow := make([]byte, 8)
				binary.BigEndian.PutUint64(bRow, row)
				key := bytes.Join([][]byte{bRow, bCol}, []byte{})
				item, err := txn.Get(key)
				if err == nil {
					item.Value(func(val []byte) error {
						o = append(o, string(val))
						return nil
					})
				}
			}
			return nil
		})
		res := make(map[string]any, len(columns))
		for i := 0; i < len(o); i++ {
			res[columns[i]] = o[i]
		}
		out <- res
	}

	db.Close()
	return nil
}
*/

type pebbleBulkWrite struct {
	db              *pebble.DB
	batch           *pebble.Batch
	highest, lowest []byte
	curSize         int
}

const (
	maxWriterBuffer = 3 << 30
)

func copyBytes(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

func (pbw *pebbleBulkWrite) Set(id []byte, val []byte) error {
	pbw.curSize += len(id) + len(val)
	if pbw.highest == nil || bytes.Compare(id, pbw.highest) > 0 {
		pbw.highest = copyBytes(id)
	}
	if pbw.lowest == nil || bytes.Compare(id, pbw.lowest) < 0 {
		pbw.lowest = copyBytes(id)
	}
	err := pbw.batch.Set(id, val, nil)
	if pbw.curSize > maxWriterBuffer {
		pbw.batch.Commit(nil)
		pbw.batch.Reset()
		pbw.curSize = 0
	}
	return err
}

func (pbw *pebbleBulkWrite) Close() {
	pbw.batch.Commit(nil)
	pbw.batch.Close()
	if pbw.lowest != nil && pbw.highest != nil {
		pbw.db.Compact(pbw.lowest, pbw.highest, true)
	}
}

func transposeOnDisk(workdir string, c csvReader, out chan map[string]any) error {

	db, err := pebble.Open(filepath.Join(workdir, "transpose.db"), &pebble.Options{})
	if err != nil {
		return err
	}

	batch := db.NewBatch()
	ptx := &pebbleBulkWrite{db, batch, nil, nil, 0}

	r, err := c.open()
	if err != nil {
		return nil
	}
	rowCount := uint64(0)
	colCount := uint64(0)
	for row := uint64(0); ; row++ {
		if (row % 100) == 0 {
			log.Printf("Row: %d", row)
		}
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if row == 0 {
			colCount = uint64(len(record))
		}
		bRow := make([]byte, 8)
		binary.BigEndian.PutUint64(bRow, row)

		for col := uint64(0); col < uint64(len(record)); col++ {
			bCol := make([]byte, 8)
			binary.BigEndian.PutUint64(bCol, col)
			//key := bytes.Join([][]byte{[]byte(bCol), bRow}, []byte{})
			key := bytes.Join([][]byte{bRow, bCol}, []byte{})
			ptx.Set(key, []byte(record[col]))
		}
		rowCount = row + 1
	}
	batch.Close()

	log.Printf("Col/Row counts: %d %d", colCount, rowCount)

	columns := []string{}

	bCol := make([]byte, 8)
	binary.BigEndian.PutUint64(bCol, 0)
	for row := uint64(0); row < rowCount; row++ {
		bRow := make([]byte, 8)
		binary.BigEndian.PutUint64(bRow, row)
		key := bytes.Join([][]byte{bRow, bCol}, []byte{})
		val, c, err := db.Get(key)
		if err == nil {
			columns = append(columns, string(val))
			c.Close()
		}
	}

	log.Printf("Columns: %#v", columns)

	for col := uint64(1); col < colCount; col++ {
		log.Printf("Writing Row %d", col)
		bCol := make([]byte, 8)
		binary.BigEndian.PutUint64(bCol, col)
		o := []string{}
		for row := uint64(0); row < rowCount; row++ {
			bRow := make([]byte, 8)
			binary.BigEndian.PutUint64(bRow, row)
			key := bytes.Join([][]byte{bRow, bCol}, []byte{})
			val, c, err := db.Get(key)
			if err == nil {
				o = append(o, string(val))
				c.Close()
			}
		}
		res := make(map[string]any, len(columns))
		for i := 0; i < len(o); i++ {
			res[columns[i]] = o[i]
		}
		out <- res
	}
	close(out)
	db.Close()
	return nil
}

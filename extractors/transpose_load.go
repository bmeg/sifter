package extractors

import (
	"bufio"
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
)

type TransposeLoadStep struct {
	Input    string `json:"input" jsonschema_description:"TSV to be transformed"`
	RowSkip  int    `json:"rowSkip" jsonschema_description:"Number of header rows to skip"`
	Sep      string `json:"sep" jsonschema_description:"Separator \\t for TSVs or , for CSVs"`
	UseDB    bool   `json:"useDB" jsonschema_description:"Do transpose without caching matrix in memory. Takes longer but works on large files"`
	UseTable int    `json:"useTable"`
}

func (ml *TransposeLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}
	inputPath, _ := task.AbsPath(input)
	cr := csvReader{inputPath, ml.RowSkip, ml.Sep}
	out := make(chan map[string]any, 10)

	if ml.UseDB {
		tdir := task.TempDir()
		go transposeInDB(tdir, cr, out)
	} else if ml.UseTable > 0 {
		tdir := task.TempDir()
		go transposeInTable(tdir, ml.UseTable, cr, out)
	} else {
		go transposeInMem(cr, out)
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
	//log.Printf("Setting %x %s", id, val)
	pbw.curSize += len(id) + len(val)
	if pbw.highest == nil || bytes.Compare(id, pbw.highest) > 0 {
		pbw.highest = copyBytes(id)
	}
	if pbw.lowest == nil || bytes.Compare(id, pbw.lowest) < 0 {
		pbw.lowest = copyBytes(id)
	}
	err := pbw.batch.Set(id, val, nil)
	if pbw.curSize > maxWriterBuffer {
		log.Printf("Running batch Commit")
		pbw.batch.Commit(nil)
		pbw.batch.Reset()
		pbw.curSize = 0
	}
	return err
}

func (pbw *pebbleBulkWrite) Close() error {
	log.Printf("Running batch close Commit")
	err := pbw.batch.Commit(nil)
	if err != nil {
		return err
	}
	pbw.batch.Close()
	if pbw.lowest != nil && pbw.highest != nil {
		pbw.db.Compact(pbw.lowest, pbw.highest, true)
	}
	return nil
}

func transposeInDB(workdir string, c csvReader, out chan map[string]any) error {

	db, err := pebble.Open(filepath.Join(workdir, "transpose.db"), &pebble.Options{})
	if err != nil {
		return err
	}

	batch := db.NewBatch()
	pbw := &pebbleBulkWrite{db, batch, nil, nil, 0}

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
			//key := bytes.Join([][]byte{bRow, bCol}, []byte{})
			//log.Printf("Put %x", key)
			key := bytes.Join([][]byte{bCol, bRow}, []byte{})
			err := pbw.Set(key, []byte(record[col]))
			if err != nil {
				log.Printf("Put Error: %s", err)
			}
		}
		rowCount = row + 1
	}
	if err := pbw.Close(); err != nil {
		log.Print(err)
	}

	log.Println(db.Metrics().String())

	log.Printf("Col/Row counts: %d %d", colCount, rowCount)

	columns := []string{}

	bCol := make([]byte, 8)
	binary.BigEndian.PutUint64(bCol, 0)
	for row := uint64(0); row < rowCount; row++ {
		bRow := make([]byte, 8)
		binary.BigEndian.PutUint64(bRow, row)
		//key := bytes.Join([][]byte{bRow, bCol}, []byte{})
		key := bytes.Join([][]byte{bCol, bRow}, []byte{})
		val, c, err := db.Get(key)
		if err == nil {
			columns = append(columns, string(val))
			c.Close()
		} else {
			log.Printf("Column error: %s", err)
		}
	}

	//log.Printf("Columns: %#v", columns)

	for col := uint64(1); col < colCount; col++ {
		if (col % 100) == 0 {
			log.Printf("Writing Col %d", col)
		}
		prefix := make([]byte, 8)
		binary.BigEndian.PutUint64(prefix, col)
		it := db.NewIter(&pebble.IterOptions{LowerBound: prefix})
		o := []string{}
		for it.First(); it.Valid() && bytes.HasPrefix(it.Key(), prefix); it.Next() {
			v := it.Value()
			r := make([]byte, len(v))
			copy(r, v)
			o = append(o, string(r))
		}
		it.Close()
		//log.Printf("Col width: %d %d", len(columns), len(o))
		if len(o) == len(columns) {
			res := make(map[string]any, len(columns))
			for i := 0; i < len(o); i++ {
				res[columns[i]] = o[i]
			}
			out <- res
		}
	}

	close(out)
	db.Close()
	os.RemoveAll(workdir)
	return nil
}

func transposeInTable(workdir string, fieldSize int, c csvReader, out chan map[string]any) error {

	table, err := os.Create(filepath.Join(workdir, "transpose"))
	if err != nil {
		return err
	}

	r, err := c.open()
	if err != nil {
		return nil
	}

	rowCount := int64(0)
	colCount := int64(0)
	writer := bufio.NewWriterSize(table, 1024*10)
	for row := int64(0); ; row++ {
		if (row % 100) == 0 {
			log.Printf("Row: %d", row)
		}
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if row == 0 {
			colCount = int64(len(record))
		}
		if colCount != int64(len(record)) {
			log.Printf("Incorrectly sized row: %d != %d", colCount, uint64(len(record)))
		}
		for col := int64(0); col < colCount; col++ {
			b := []byte(record[col])
			rec := make([]byte, fieldSize)
			copy(rec, b)
			writer.Write(rec)
		}
		rowCount = row + 1
	}
	writer.Flush()

	stepSize := colCount * int64(fieldSize)

	log.Printf("Col/Row counts: %d %d", colCount, rowCount)

	columns := []string{}
	for row := int64(0); row < rowCount; row++ {
		buf := make([]byte, fieldSize)
		table.ReadAt(buf, row*stepSize)
		tmp := bytes.Split(buf, []byte{0})
		if err == nil {
			columns = append(columns, string(tmp[0]))
		} else {
			log.Printf("Column error: %s", err)
		}
	}

	log.Printf("Columns: %s\n", columns)

	for col := int64(1); col < colCount; col++ {
		if (col % 100) == 0 {
			log.Printf("Writing Col %d", col)
		}
		record := map[string]any{}
		for row := int64(0); row < rowCount; row++ {
			buf := make([]byte, fieldSize)
			table.ReadAt(buf, row*stepSize)
			tmp := bytes.Split(buf, []byte{0})
			record[columns[row]] = string(tmp[0])
		}
		out <- record
	}
	table.Close()
	close(out)
	os.RemoveAll(workdir)
	return nil
}

package extractors

import (
	"io"
	"log"
	"os"
	"strings"

	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/csv"

	//"encoding/gob"

	"path/filepath"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/manager"

	"github.com/dgraph-io/badger/v2"
)

type TransposeFileStep struct {
	Input    string `json:"input" jsonschema_description:"TSV to transpose"`
	Output   string `json:"output" jsonschema_description:"Where transpose output should be stored"`
	LineSkip int    `json:"lineSkip" jsonschema_description:"Number of header lines to skip"`
	LowMem   bool   `json:"lowMem" jsonschema_description:"Do transpose without caching matrix in memory. Takes longer but works on large files"`
}

type csvReader struct {
	inputPath string
	lineSkip  int
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
	r.FieldsPerRecord = -1
	for i := 0; i < c.lineSkip; i++ {
		r.Read()
	}
	return r, nil
}

func transposeInMem(c csvReader, out *csv.Writer) error {
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
	for i := 0; i < l; i++ {
		o := make([]string, h)
		for j := 0; j < h; j++ {
			o[j] = matrix[j][i]
		}
		out.Write(o)
	}
	return nil
}

func transposeOnDisk(workdir string, c csvReader, out *csv.Writer) error {
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

	for col := uint64(0); col < colCount; col++ {
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
		out.Write(o)
	}

	db.Close()
	return nil
}

func (ml *TransposeFileStep) Run(task *manager.Task) error {

	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
	output, err := evaluate.ExpressionString(ml.Output, task.Inputs, nil)

	inputPath, err := task.AbsPath(input)
	outputPath, err := task.AbsPath(output)

	cr := csvReader{inputPath, ml.LineSkip}
	ohd, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	w := csv.NewWriter(ohd)
	w.Comma = '\t'

	if !ml.LowMem {
		transposeInMem(cr, w)
		log.Printf("Transpose Done: %s", outputPath)
	} else {
		tdir := task.TempDir()
		transposeOnDisk(tdir, cr, w)
		log.Printf("Transpose Done: %s", outputPath)
	}
	w.Flush()
	ohd.Close()
	return nil
}

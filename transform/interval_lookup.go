package transform

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type TSVInterval struct {
	Input  string   `json:"input"`
	Sep    string   `json:"sep"`
	Value  string   `json:"value"`
	Key    string   `json:"sequence"`
	Start  string   `json:"start"`
	Stop   string   `json:"stop"`
	Header []string `json:"header"`
}

type IntervalLookupStep struct {
	Sequence string       `json:"sequence"`
	Position string       `json:"position"`
	TSV      *TSVInterval `json:"tsv"`
	Dest     string       `json:"dest"`
}

type intervalTable interface {
	LookupValue(seq string, pos int) ([]string, bool)
	LookupRecord(seq string, pos int) ([]map[string]any, bool)
}

type intervalProcess struct {
	config     *IntervalLookupStep
	table      intervalTable
	userConfig map[string]string
	hitCount   int
	missCount  int
}

func (ils *IntervalLookupStep) Init(task task.RuntimeTask) (Processor, error) {

	if ils.TSV != nil {
		var table intervalTable
		var err error
		if table, err = ils.TSV.open(task); err == nil {
			return &intervalProcess{ils, table, task.GetConfig(), 0, 0}, nil
		}
		return nil, err
	}

	return nil, fmt.Errorf("table input not defined")

}

func (tp *intervalProcess) Close() {
	log.Printf("Table hits: %d misses: %d", tp.hitCount, tp.missCount)
}

func (tp *intervalProcess) Process(row map[string]any) []map[string]any {
	seq, err := evaluate.ExpressionString(tp.config.Sequence, tp.userConfig, row)
	if err != nil {
		return []map[string]any{row}
	}

	position, err := evaluate.ExpressionString(tp.config.Position, tp.userConfig, row)
	if err != nil {
		return []map[string]any{row}
	}

	positionInt, err := strconv.Atoi(position)
	if err != nil {
		return []map[string]any{row}
	}
	data, ok := tp.table.LookupRecord(seq, positionInt)

	if ok {
		row[tp.config.Dest] = data
	}
	return []map[string]any{row}
}

func (tsv *TSVInterval) open(task task.RuntimeTask) (intervalTable, error) {
	inputPath, err := evaluate.ExpressionString(tsv.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	if s, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", inputPath)
	} else if s.IsDir() {
		return nil, fmt.Errorf("file not found: %s", inputPath)
	}
	log.Printf("Loading Translation file: %s", inputPath)

	var inputStream io.Reader
	if gfile, err := os.Open(inputPath); err == nil {
		if strings.HasSuffix(inputPath, ".gz") {
			inp, err := gzip.NewReader(gfile)
			if err != nil {
				return nil, err
			}
			inputStream = inp
		} else {
			inputStream = gfile
		}
	} else if err != nil {
		log.Printf("Error loading table: %s", err)
		return nil, err
	}

	if tsv.Sep == "" {
		tsv.Sep = "\t"
	}

	tsvReader := csv.NewReader(inputStream)
	tsvReader.Comma = rune(tsv.Sep[0])

	tp := &tsvIntervalLookup{config: tsv, inputs: task.GetConfig()}

	lines, err := tsvReader.ReadAll()
	if err != nil {
		log.Printf("Error loading lookup table: %s", err)
		return nil, err
	}

	for _, row := range lines {
		if tp.colmap == nil {
			tp.colmap = map[string]int{}
			for i, k := range row {
				tp.colmap[k] = i
			}
		} else {
			tp.table[row[tp.colmap[tsv.Key]]] = row
		}
	}
	log.Printf("tableLookup loaded %d values from %s", len(tp.table), inputPath)
	return tp, nil

}

type tsvIntervalLookup struct {
	config *TSVInterval
	inputs map[string]string
}

func (til *tsvIntervalLookup) LookupValue(seq string, pos int) ([]string, bool) {

}

func (til *tsvIntervalLookup) LookupRecord(seq string, pos int) ([]map[string]any, bool) {

}

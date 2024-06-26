package transform

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"encoding/csv"
	"encoding/json"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"

	"github.com/bmeg/sifter/evaluate"
)

type TSVTable struct {
	Input  string   `json:"input"`
	Sep    string   `json:"sep"`
	Value  string   `json:"value"`
	Key    string   `json:"key"`
	Header []string `json:"header"`
}

type InlineTable map[string]string

type JSONTable struct {
	Input string `json:"input"`
	Value string `json:"value"`
	Key   string `json:"key"`
}

type PipelineTable struct {
	From  string `json:"from"`
	Value string `json:"value"`
	Key   string `json:"key"`
}

type lookupTable interface {
	LookupValue(k string) (string, bool)
	LookupRecord(k string) (map[string]any, bool)
}

type jsonLookup struct {
	config *JSONTable
	inputs map[string]string
	table  map[string][]byte //found it more space efficiant to store the JSON rather then keep all the unpacked values
}

type tsvLookup struct {
	config *TSVTable
	inputs map[string]string
	colmap map[string]int
	table  map[string][]string
}

type LookupStep struct {
	Replace  string            `json:"replace"`
	TSV      *TSVTable         `json:"tsv"`
	JSON     *JSONTable        `json:"json"`
	Table    *LookupTable      `json:"table"`
	Pipeline *PipelineTable    `json:"pipeline"`
	Lookup   string            `json:"lookup"`
	Copy     map[string]string `json:"copy"`
	//Mapping map[string]string `json:"mapping"`
}

type lookupProcess struct {
	config     *LookupStep
	table      lookupTable
	userConfig map[string]string
	hitCount   int
	missCount  int
	//table  map[string][]string
}

func (tr *LookupStep) Init(task task.RuntimeTask) (Processor, error) {
	if tr.TSV != nil {
		var table lookupTable
		var err error
		if table, err = tr.TSV.open(task); err == nil {
			return &lookupProcess{tr, table, task.GetConfig(), 0, 0}, nil
		}
		return nil, err
	} else if tr.JSON != nil {
		var table lookupTable
		var err error
		if table, err = tr.JSON.open(task); err == nil {
			return &lookupProcess{tr, table, task.GetConfig(), 0, 0}, nil
		}
		return nil, err
	} else if tr.Table != nil {
		return &lookupProcess{tr, tr.Table, task.GetConfig(), 0, 0}, nil
	} else if tr.Pipeline != nil {
		return &pipelineLookupProcess{tr, task}, nil
	}
	return nil, fmt.Errorf("table input not defined")
}

func (tr *LookupStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	if tr.TSV != nil && tr.TSV.Input != "" {
		for _, s := range evaluate.ExpressionIDs(tr.TSV.Input) {
			out = append(out, config.Variable{Type: config.File, Name: config.TrimPrefix(s)})
		}
	} else if tr.JSON != nil && tr.JSON.Input != "" {
		for _, s := range evaluate.ExpressionIDs(tr.JSON.Input) {
			out = append(out, config.Variable{Type: config.File, Name: config.TrimPrefix(s)})
		}
	}
	return out
}

func (tp *lookupProcess) Close() {
	logger.Info("Table summary", "hit", tp.hitCount, "miss", tp.missCount)
}

func (tp *lookupProcess) Process(row map[string]interface{}) []map[string]interface{} {
	if tp.config.Replace != "" {
		if _, ok := row[tp.config.Replace]; ok {
			out := map[string]interface{}{}
			for k, v := range row {
				if k == tp.config.Replace {
					if x, ok := v.(string); ok {
						if tp.config.Copy != nil {
							if n, ok := tp.table.LookupRecord(x); ok {
								t := map[string]any{}
								for tk, tv := range tp.config.Copy {
									if jv, ok := n[tv]; ok {
										t[tk] = jv
									}
								}
								out[k] = t
							} else {
								out[k] = x
							}
						} else {
							if n, ok := tp.table.LookupValue(x); ok {
								out[k] = n
							} else {
								out[k] = x
							}
						}
					} else if x, ok := v.([]interface{}); ok {
						o := []interface{}{}
						for _, y := range x {
							if z, ok := y.(string); ok {
								if tp.config.Copy != nil {
									if n, ok := tp.table.LookupRecord(z); ok {
										t := map[string]any{}
										for tk, tv := range tp.config.Copy {
											if jv, ok := n[tv]; ok {
												t[tk] = jv
											}
										}
										o = append(o, t)
									} else {
										o = append(o, z)
									}
								} else {
									if n, ok := tp.table.LookupValue(z); ok {
										o = append(o, n)
									} else {
										o = append(o, z)
									}
								}
							}
						}
						out[k] = o
					} else if x, ok := v.(map[string]any); ok {
						o := map[string]any{}
						for key, val := range x {
							if n, ok := tp.table.LookupValue(key); ok {
								o[n] = val
							} else {
								o[key] = val
							}
						}
						out[k] = o
					} else {
						out[k] = v
					}
				} else {
					out[k] = v
				}
			}
			return []map[string]any{out}
		}
	} else if tp.config.Lookup != "" {
		value, err := evaluate.ExpressionString(tp.config.Lookup, tp.userConfig, row)
		if err == nil {
			out := map[string]any{}
			for k, v := range row {
				out[k] = v
			}
			if pv, ok := tp.table.LookupRecord(value); ok {
				for k, v := range tp.config.Copy {
					if ki, ok := pv[v]; ok {
						out[k] = ki
						tp.hitCount++
					}
				}
			} else {
				tp.missCount++
			}
			return []map[string]any{out}
		}
	}
	return []map[string]any{row}
}

func (tsv *TSVTable) open(task task.RuntimeTask) (lookupTable, error) {
	inputPath, err := evaluate.ExpressionString(tsv.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	if s, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", inputPath)
	} else if s.IsDir() {
		return nil, fmt.Errorf("file not found: %s", inputPath)
	}
	logger.Debug("Loading Translation file", "path", inputPath)

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
		logger.Error("Error loading table", "error", err)
		return nil, err
	}

	if tsv.Sep == "" {
		tsv.Sep = "\t"
	}

	tsvReader := csv.NewReader(inputStream)
	tsvReader.Comma = rune(tsv.Sep[0])

	tp := &tsvLookup{config: tsv, inputs: task.GetConfig()}

	tp.colmap = nil
	if len(tsv.Header) > 0 {
		tp.colmap = map[string]int{}
		for i, n := range tsv.Header {
			tp.colmap[n] = i
		}
	}
	tp.table = map[string][]string{}

	lines, err := tsvReader.ReadAll()
	if err != nil {
		logger.Error("Error loading lookup table", "error", err)
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
	logger.Debug("tableLookup summary", "count", len(tp.table), "path", inputPath)
	return tp, nil
}

func (tl *tsvLookup) LookupValue(w string) (string, bool) {
	c := tl.colmap[tl.config.Value]
	if a, ok := tl.table[w]; ok {
		return a[c], true
	}
	return "", false
}

func (tl *tsvLookup) LookupRecord(w string) (map[string]any, bool) {
	if a, ok := tl.table[w]; ok {
		out := map[string]any{}
		for c, i := range tl.colmap {
			out[c] = a[i]
		}
		return out, true
	}
	return nil, false
}

func (jf *JSONTable) open(task task.RuntimeTask) (lookupTable, error) {
	inputPath, err := evaluate.ExpressionString(jf.Input, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	if s, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", inputPath)
	} else if s.IsDir() {
		return nil, fmt.Errorf("file not found: %s", inputPath)
	}
	logger.Debug("Loading Translation file", "path", inputPath)

	var inputStream chan []byte
	if strings.HasSuffix(inputPath, ".gz") {
		inputStream, err = golib.ReadGzipLines(inputPath)
	} else {
		inputStream, err = golib.ReadFileLines(inputPath)
	}
	if err != nil {
		return nil, err
	}

	jp := &jsonLookup{jf, task.GetConfig(), map[string][]byte{}}
	for line := range inputStream {
		if len(line) > 0 {
			row := map[string]interface{}{}
			err := json.Unmarshal(line, &row)
			if err != nil {
				return nil, err
			}
			if key, ok := row[jf.Key]; ok {
				if keyStr, ok := key.(string); ok {
					jp.table[keyStr] = line
				}
			}
		}
	}
	logger.Debug("jsonLookup summary", "count", len(jp.table), "path", inputPath)

	return jp, nil
}

func (jp *jsonLookup) LookupValue(s string) (string, bool) {
	if line, ok := jp.table[s]; ok {
		row := map[string]interface{}{}
		json.Unmarshal(line, &row)
		if x, ok := row[jp.config.Value]; ok {
			if xStr, ok := x.(string); ok {
				return xStr, true
			}
		}
	}
	return "", false
}

func (jp *jsonLookup) LookupRecord(s string) (map[string]any, bool) {
	if line, ok := jp.table[s]; ok {
		row := map[string]interface{}{}
		json.Unmarshal(line, &row)
		return row, true
	}
	return nil, false
}

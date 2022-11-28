package transform

import (
	"fmt"
	"log"
	"os"
	"strings"

	"encoding/json"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/config"
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

type lookupTable interface {
	LookupValue(k string) (string, bool)
	LookupRecord(k string) (map[string]any, bool)
}

type LookupTable map[string]string

type LookupStep struct {
	Replace string            `json:"replace"`
	TSV     *TSVTable         `json:"tsv"`
	JSON    *JSONTable        `json:"json"`
	Table   *LookupTable      `json:"table"`
	Lookup  string            `json:"lookup"`
	Copy    map[string]string `json:"copy"`
	//Mapping map[string]string `json:"mapping"`
}

type lookupProcess struct {
	config     *LookupStep
	table      lookupTable
	userConfig map[string]string
	//table  map[string][]string
}

func (tr *LookupStep) Init(task task.RuntimeTask) (Processor, error) {
	if tr.TSV != nil {
		var table lookupTable
		var err error
		if table, err = tr.TSV.open(task); err == nil {
			return &lookupProcess{tr, table, task.GetConfig()}, nil
		}
		return nil, err
	} else if tr.JSON != nil {
		var table lookupTable
		var err error
		if table, err = tr.JSON.open(task); err == nil {
			return &lookupProcess{tr, table, task.GetConfig()}, nil
		}
		return nil, err
	} else if tr.Table != nil {
		return &lookupProcess{tr, tr.Table, task.GetConfig()}, nil
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

func (tp *lookupProcess) Close() {}

func (tp *lookupProcess) Process(i map[string]interface{}) []map[string]interface{} {
	if tp.config.Replace != "" {
		if _, ok := i[tp.config.Replace]; ok {
			out := map[string]interface{}{}
			for k, v := range i {
				if k == tp.config.Replace {
					if x, ok := v.(string); ok {
						if n, ok := tp.table.LookupValue(x); ok {
							out[k] = n
						} else {
							out[k] = x
						}
					} else if x, ok := v.([]interface{}); ok {
						o := []interface{}{}
						for _, y := range x {
							if z, ok := y.(string); ok {
								if n, ok := tp.table.LookupValue(z); ok {
									o = append(o, n)
								} else {
									o = append(o, z)
								}
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
		value, err := evaluate.ExpressionString(tp.config.Lookup, tp.userConfig, i)
		if err == nil {
			if pv, ok := tp.table.LookupRecord(value); ok {
				for k, v := range tp.config.Copy {
					if ki, ok := pv[v]; ok {
						i[k] = ki
					}
				}
			}
		}
	}
	return []map[string]any{i}
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
	log.Printf("Loading Translation file: %s", inputPath)

	var inputStream chan []byte
	if strings.HasSuffix(inputPath, ".gz") {
		inputStream, err = golib.ReadGzipLines(inputPath)
	} else {
		inputStream, err = golib.ReadFileLines(inputPath)
	}
	if err != nil {
		return nil, err
	}

	if tsv.Sep == "" {
		tsv.Sep = "\t"
	}

	tp := &tsvLookup{config: tsv, inputs: task.GetConfig()}

	tp.colmap = nil
	if len(tsv.Header) > 0 {
		tp.colmap = map[string]int{}
		for i, n := range tsv.Header {
			tp.colmap[n] = i
		}
	}
	tp.table = map[string][]string{}
	for line := range inputStream {
		if len(line) > 0 {
			row := strings.Split(string(line), tsv.Sep)
			if tp.colmap == nil {
				tp.colmap = map[string]int{}
				for i, k := range row {
					tp.colmap[k] = i
				}
			} else {
				tp.table[row[tp.colmap[tsv.Key]]] = row
			}
		}
	}
	log.Printf("tableLookup loaded %d values from %s", len(tp.table), inputPath)
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
	log.Printf("Loading Translation file: %s", inputPath)

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
	log.Printf("jsonLookup loaded %d values from %s", len(jp.table), inputPath)

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

func (jp *LookupTable) LookupValue(k string) (string, bool) {
	s, ok := (*jp)[k]
	return s, ok
}

func (jp *LookupTable) LookupRecord(k string) (map[string]any, bool) {
	if x, ok := (*jp)[k]; ok {
		return map[string]any{"value": x}, true
	}
	return nil, false
}

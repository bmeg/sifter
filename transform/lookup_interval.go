package transform

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"

	"github.com/rdleal/intervalst/interval"
)

type NamedPosition struct {
	Name string
	Pos  int64
}

func NamedCmp(x, y NamedPosition) int {
	v := strings.Compare(x.Name, y.Name)
	if v != 0 {
		return v
	}
	return int(x.Pos - y.Pos)
}

func NumCmp(x, y int64) int {
	return int(x - y)
}

type IntervalStep struct {
	Match string        `json:"match"`
	Start string        `json:"start"`
	End   string        `json:"end"`
	Field string        `json:"field"`
	JSON  *JSONInterval `json:"json"`
	//TSV   *TSVInterval      `json:"tsv"`
}

type JSONInterval struct {
	Match string `json:"match"`
	Input string `json:"input"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type TSVInterval struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type intervalProcess struct {
	config    *IntervalStep
	table     map[string]*interval.SearchTree[map[string]any, int64]
	hitCount  int
	missCount int
	//table  map[string][]string
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int8:
		return int64(t)
	case int16:
		return int64(t)
	case int32:
		return int64(t)
	case int64:
		return int64(t)
	case uint:
		return int64(t)
	case uint8:
		return int64(t)
	case uint16:
		return int64(t)
	case uint32:
		return int64(t)
	case uint64:
		return int64(t)
	case float32:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err == nil {
			return i
		}
	}
	//fmt.Printf("Issues: %T\n", v)
	return 0
}

func (tr *IntervalStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	if tr.JSON != nil && tr.JSON.Input != "" {
		for _, s := range evaluate.ExpressionIDs(tr.JSON.Input) {
			out = append(out, config.Variable{Type: config.File, Name: config.TrimPrefix(s)})
		}
	}
	return out
}

func (tr *IntervalStep) Init(task task.RuntimeTask) (Processor, error) {
	if tr.JSON != nil {

		tableMap := map[string]*interval.SearchTree[map[string]any, int64]{}
		var err error

		inputPath, err := evaluate.ExpressionString(tr.JSON.Input, task.GetConfig(), nil)
		logger.Debug("Opening interval file", "path", inputPath, "input", tr.JSON.Input, "config", task.GetConfig())
		if err != nil {
			return nil, err
		}

		var inputStream chan []byte
		if strings.HasSuffix(inputPath, ".gz") {
			inputStream, err = golib.ReadGzipLines(inputPath)
		} else {
			inputStream, err = golib.ReadFileLines(inputPath)
		}
		if err != nil {
			return nil, err
		}
		count := 0
		for line := range inputStream {
			if len(line) > 0 {
				row := map[string]any{}
				err := json.Unmarshal(line, &row)
				if err != nil {
					return nil, err
				}
				if match, ok := row[tr.JSON.Match]; ok {
					if matchString, ok := match.(string); ok {
						var table *interval.SearchTree[map[string]any, int64]
						if table, ok = tableMap[matchString]; !ok {
							table = interval.NewSearchTree[map[string]any](NumCmp)
							tableMap[matchString] = table
						}

						if startString, ok := row[tr.JSON.Start]; ok {
							start := toInt64(startString)
							if endString, ok := row[tr.JSON.End]; ok {
								end := toInt64(endString)
								if count < 10 {
									logger.Debug("intervalIntersect load", "match", matchString, "start", start, "end", end)
								}
								table.Insert(start, end, row)
								count++
							}
						}
					}
				}
			}
		}

		logger.Debug("intervalIntersect loaded", "values", count)

		return &intervalProcess{config: tr, table: tableMap}, err
	}
	return nil, fmt.Errorf("interval input not defined")
}

func (tp *intervalProcess) Close() {
	logger.Info("Table summary", "hit", tp.hitCount, "miss", tp.missCount)
}

func (tp *intervalProcess) Process(row map[string]interface{}) []map[string]interface{} {

	if match, ok := row[tp.config.Match]; ok {
		if matchString, ok := match.(string); ok {
			if table, ok := tp.table[matchString]; ok {
				if start, ok := row[tp.config.Start]; ok {
					startInt := toInt64(start)
					if end, ok := row[tp.config.End]; ok {
						endInt := toInt64(end)
						vals, found := table.AllIntersections(
							startInt,
							endInt,
						)
						if found {
							row[tp.config.Field] = vals
							tp.hitCount++
						} else {
							if tp.missCount < 10 {
								logger.Debug("intervalIntersect miss",
									"match", matchString, "start", startInt, "end", endInt)
							}
							row[tp.config.Field] = []any{}
							tp.missCount++
						}
					}
				}
			}
		}
	}

	return []map[string]any{row}
}

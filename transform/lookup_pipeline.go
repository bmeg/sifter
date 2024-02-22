package transform

import (
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type pipelineLookupProcess struct {
	config *LookupStep
	task   task.RuntimeTask
}

func (pls *pipelineLookupProcess) GetRightPipeline() string {
	return pls.config.Pipeline.From
}

func (pls *pipelineLookupProcess) Process(left chan map[string]any, right chan map[string]any, out chan map[string]any) {
	//lookup table comes from the right, so consume that first

	vals := map[string]map[string]any{}
	for i := range right {
		if key, ok := i[pls.config.Pipeline.Key]; ok {
			if kStr, ok := key.(string); ok {
				vals[kStr] = i
			}
		} else {
			logger.Info("Missing key %s : %#v", pls.config.Pipeline.Key, i)
		}
	}
	logger.Info("Pipeline lookup loaded %d values", len(vals))
	table := recordTable{vals, pls.config.Pipeline.Value}
	lk := lookupProcess{pls.config, &table, pls.task.GetConfig(), 0, 0}

	for i := range left {
		o := lk.Process(i)
		out <- o[0]
	}

	lk.Close()

}

func (pls *pipelineLookupProcess) Close() {

}

type recordTable struct {
	table      map[string]map[string]any
	valueField string
}

func (jp *recordTable) LookupValue(k string) (string, bool) {
	if v, ok := jp.table[k]; ok {
		if s, ok := v[jp.valueField]; ok {
			str, ok := s.(string)
			return str, ok
		}
	}
	return "", false
}

func (jp *recordTable) LookupRecord(k string) (map[string]any, bool) {
	v, ok := jp.table[k]
	return v, ok
}

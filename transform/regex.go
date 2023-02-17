package transform

import (
	"regexp"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type RegexReplaceStep struct {
	Field   string `json:"field"`
	Regex   string `json:"regex"`
	Replace string `json:"replace"`
	Dest    string `json:"dst"`
}

type regexReplaceProcess struct {
	config *RegexReplaceStep
	task   task.RuntimeTask
	reg    *regexp.Regexp
}

func (pr *RegexReplaceStep) Init(t task.RuntimeTask) (Processor, error) {
	reg, err := regexp.Compile(pr.Regex)
	if err != nil {
		return nil, err
	}
	return &regexReplaceProcess{pr, t, reg}, nil
}

func (re *regexReplaceProcess) Close() {}

func (re *regexReplaceProcess) PoolReady() bool {
	return true
}

func (re *regexReplaceProcess) Process(i map[string]interface{}) []map[string]interface{} {
	if field, ok := i[re.config.Field]; ok {
		if fStr, ok := field.(string); ok {
			replace, _ := evaluate.ExpressionString(re.config.Replace, re.task.GetConfig(), i)
			dst := re.config.Field
			if re.config.Dest != "" {
				dst = re.config.Dest
			}
			o := re.reg.ReplaceAllString(fStr, replace)
			z := map[string]interface{}{}
			for x, y := range i {
				z[x] = y
			}
			z[dst] = o
			return []map[string]any{z}
		}
	}
	return []map[string]any{i}
}

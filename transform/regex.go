package transform

import (
	"regexp"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/pipeline"
)

type RegexReplaceStep struct {
	Column  string `json:"col"`
	Regex   string `json:"regex"`
	Replace string `json:"replace"`
	Dest    string `json:"dst"`
	reg     *regexp.Regexp
}

func (re RegexReplaceStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	col, _ := evaluate.ExpressionString(re.Column, task.Inputs, i)
	replace, _ := evaluate.ExpressionString(re.Replace, task.Inputs, i)
	dst, _ := evaluate.ExpressionString(re.Dest, task.Inputs, i)

	o := re.reg.ReplaceAllString(col, replace)
	z := map[string]interface{}{}
	for x, y := range i {
		z[x] = y
	}
	z[dst] = o
	return z
}

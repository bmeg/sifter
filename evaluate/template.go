package evaluate

import (
	//"fmt"
	//"regexp"
	"strings"

	"github.com/aymerick/raymond"
)

func init() {
	raymond.RegisterHelper("split-select", func(in, split string, i int) string {
		o := strings.Split(in, split)
		if i >= 0 && i < len(o) {
			return o[i]
		}
		return in
	})
}

func ExpressionString(expression string, config map[string]interface{}, row map[string]interface{}) (string, error) {
	d := map[string]interface{}{"config": config}
	if row != nil {
		d["row"] = row
	}
	return raymond.Render(expression, d)
}

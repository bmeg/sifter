package evaluate

import (
	//"fmt"
	//"regexp"
	"github.com/aymerick/raymond"
)

func ExpressionString(expression string, inputs map[string]interface{}, row map[string]interface{}) (string, error) {
	d := map[string]interface{}{"inputs":inputs}
	if row != nil {
		d["row"] = row
	}
	return raymond.Render(expression, d)
}

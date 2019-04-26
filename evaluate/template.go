package evaluate

import (
	//"fmt"
	//"log"
	//"regexp"
	"github.com/aymerick/raymond"
)

func ExpressionString(expression string, inputs map[string]interface{}, row map[string]interface{}) (string, error) {
	d := map[string]interface{}{"input":inputs}
	if row != nil {
		d["row"] = row
	}
	return raymond.Render(expression, d)
}

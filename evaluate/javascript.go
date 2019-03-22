package evaluate

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"gopkg.in/olebedev/go-duktape.v3"
)

var EXP_RE_STRING, _ = regexp.Compile(`(.*)\$\((.*)\)(.*)`)
var EXP_RE, _ = regexp.Compile(`\$\((.*)\)`)

func ExpressionString(expression string, inputs map[string]interface{}) (string, error) {

	matches := EXP_RE_STRING.FindStringSubmatch(expression)
	if matches == nil {
		return expression, nil
	}
	code := matches[2]
	log.Printf("JS Expression: %s", code)
	//log.Printf("JS Inputs: %#v", self.Inputs.Normalize())

	ctx := duktape.New()

	inStr, _ := json.Marshal(inputs)

	if err := ctx.PevalStringNoresult(fmt.Sprintf("var input = %s;", inStr)); err != 0 {
		return "", fmt.Errorf("Input variable load failed")
	}
	if err := ctx.PevalString(code); err != nil {
		return "", fmt.Errorf("Config string failed: %s", code)
	}
	out := ctx.GetString(-1)
	return out, nil
}

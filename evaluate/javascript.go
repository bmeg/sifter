package evaluate

import (
	//"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/oliveagle/jsonpath"
)

var EXP_RE_STRING, _ = regexp.Compile(`$\.(.*)`)

func ExpressionString(expression string, inputs map[string]interface{}) (string, error) {

	matches := EXP_RE_STRING.FindStringSubmatch(expression)
	if matches == nil {
		return expression, nil
	}

	log.Printf("JSON PATH: %s", expression)

	res, err := jsonpath.JsonPathLookup(inputs, expression)
	if err != nil {
		return "", err
	}
	if out, ok := res.(string); ok {
		return out, nil
	}
	return "", fmt.Errorf("Need string variable")
}

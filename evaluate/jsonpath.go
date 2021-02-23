package evaluate

import (
	"strings"

	"github.com/bmeg/jsonpath"
)

func SetJSONPath(expression string, data map[string]interface{}, value interface{}) error {
	if !strings.HasPrefix(expression, "$.") {
		expression = "$." + expression
	}
	return jsonpath.JsonPathSet(data, expression, value)
}

func GetJSONPath(expression string, inputs map[string]interface{}) (interface{}, error) {
	return jsonpath.JsonPathLookup(inputs, "$."+expression)
}

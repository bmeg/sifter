
package evaluate

import (
    "github.com/oliveagle/jsonpath"
)


func GetJSONPath(expression string, inputs map[string]interface{}) (interface{}, error) {
  return jsonpath.JsonPathLookup(inputs, "$." + expression)
}

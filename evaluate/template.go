package evaluate

import (
	//"fmt"
	//"regexp"

	"strings"

	"github.com/aymerick/raymond"
	"github.com/aymerick/raymond/lexer"
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

func ExpressionString(expression string, config map[string]string, row map[string]interface{}) (string, error) {
	d := map[string]interface{}{"config": config}
	if row != nil {
		d["row"] = row
	}
	return raymond.Render(expression, d)
}

func ExpressionIDs(expression string) []string {
	out := []string{}
	lex := lexer.Scan(expression)
	cur := ""
	for {
		token := lex.NextToken()
		//pick up all ID tokens
		if token.Kind == lexer.TokenID {
			cur = cur + token.Val
		} else if token.Kind == lexer.TokenSep {
			cur = cur + token.Val
		} else {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
		}
		// stops when all tokens have been consumed, or on error
		if token.Kind == lexer.TokenEOF || token.Kind == lexer.TokenError {
			break
		}
	}
	return out
}

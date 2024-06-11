package evaluate

import (
	"fmt"

	"github.com/bmeg/sifter/logger"
	"github.com/go-python/gpython/py"
	_ "github.com/go-python/gpython/stdlib" // Load modiles in gpython environment
)

type GPythonEngine struct{}

type GPythonProcessor struct {
	code   *PyCode
	method string
}

func (g GPythonEngine) Compile(code string, method string) (Processor, error) {
	out := GPythonProcessor{}
	o, err := PyCompile(code)
	if err != nil {
		return out, err
	}
	out.code = o
	out.method = method
	return out, nil
}

func (g GPythonProcessor) Close() {}

func (g GPythonProcessor) Evaluate(inputs ...map[string]interface{}) (map[string]interface{}, error) {
	return g.code.Evaluate(g.method, inputs...)
}

func (g GPythonProcessor) EvaluateArray(inputs ...map[string]interface{}) ([]any, error) {
	return g.code.EvaluateArray(g.method, inputs...)
}

func (g GPythonProcessor) EvaluateBool(inputs ...map[string]interface{}) (bool, error) {
	return g.code.EvaluateBool(g.method, inputs...)
}

func PyObject(i interface{}) py.Object {
	if xMap, ok := i.(map[string]interface{}); ok {
		o := py.StringDict{}
		for k, v := range xMap {
			o[k] = PyObject(v)
		}
		return o
	} else if xList, ok := i.([]interface{}); ok {
		o := py.NewList()
		for _, v := range xList {
			o.Append(PyObject(v))
		}
		return o
	} else if xList, ok := i.([]map[string]any); ok {
		o := py.NewList()
		for _, v := range xList {
			o.Append(PyObject(v))
		}
		return o
	} else if xList, ok := i.([]string); ok {
		o := py.NewList()
		for _, v := range xList {
			o.Append(PyObject(v))
		}
		return o
	} else if xString, ok := i.(string); ok {
		return py.String(xString)
	} else if xInt, ok := i.(int); ok {
		return py.Int(xInt)
	} else if xInt, ok := i.(int64); ok {
		return py.Int(xInt)
	} else if xFloat, ok := i.(float64); ok {
		return py.Float(xFloat)
	} else if xFloat, ok := i.(float32); ok {
		return py.Float(xFloat)
	} else if xBool, ok := i.(bool); ok {
		return py.Bool(xBool)
	} else if i == nil {
		return py.None
	} else {
		logger.Error("gpython conversion Not found: %T", i)
	}
	return nil
}

func FromPyObject(i py.Object) interface{} {
	if xMap, ok := i.(py.StringDict); ok {
		out := map[string]interface{}{}
		for k, v := range xMap {
			out[k] = FromPyObject(v)
		}
		return out
	} else if xList, ok := i.(*py.List); ok {
		out := []interface{}{}
		for _, v := range xList.Items {
			out = append(out, FromPyObject(v))
		}
		return out
	} else if xGen, ok := i.(*py.Generator); ok {
		out := []any{}
		for done := false; !done; {
			i, err := xGen.M__next__()
			if err == nil {
				j := FromPyObject(i)
				out = append(out, j)
			} else {
				done = true
			}
		}
		return out
	} else if xTuple, ok := i.(py.Tuple); ok {
		out := []interface{}{}
		for _, v := range xTuple {
			out = append(out, FromPyObject(v))
		}
		return out
	} else if xStr, ok := i.(py.String); ok {
		return string(xStr)
	} else if xFloat, ok := i.(py.Float); ok {
		return float64(xFloat)
	} else if xBool, ok := i.(py.Bool); ok {
		return bool(xBool)
	} else if xInt, ok := i.(py.Int); ok {
		return int64(xInt)
	} else if i == py.None {
		return nil
	} else {
		logger.Error("gpython conversion Not found: %T", i)
	}
	return nil
}

type PyCode struct {
	module *py.Module
}

func PyCompile(codeStr string) (*PyCode, error) {

	logger.Debug("Gpython compile: %s", codeStr)

	opts := py.ContextOpts{SysArgs: []string{}, SysPaths: []string{}}
	ctx := py.NewContext(opts)

	mainImpl := py.ModuleImpl{
		Info:    py.ModuleInfo{Name: "user", FileDesc: "<user>"},
		CodeSrc: codeStr,
	}
	module, err := ctx.ModuleInit(&mainImpl)
	if err != nil {
		py.TracebackDump(err)
		return nil, err
	}

	return &PyCode{module: module}, nil
}

func (p *PyCode) Evaluate(method string, inputs ...map[string]interface{}) (map[string]interface{}, error) {
	fun := p.module.Globals[method]
	in := py.Tuple{}
	for _, i := range inputs {
		data := PyObject(i)
		in = append(in, data)
	}
	out, err := py.Call(fun, in, nil)
	if err != nil {
		py.TracebackDump(err)
		logger.Error("Error Inputs: %#v", inputs)
		logger.Error("Error Inputs: %#v", in)
		logger.Error("Map Error: %s", err)
		return nil, err
	}
	o := FromPyObject(out)
	if out, ok := o.(map[string]interface{}); ok {
		return out, nil
	}
	return nil, fmt.Errorf("incorrect return type: %s", out)
}

func (p *PyCode) EvaluateArray(method string, inputs ...map[string]interface{}) ([]any, error) {
	fun := p.module.Globals[method]
	in := py.Tuple{}
	for _, i := range inputs {
		data := PyObject(i)
		in = append(in, data)
	}
	out, err := py.Call(fun, in, nil)
	if err != nil {
		py.TracebackDump(err)
		logger.Error("Error Inputs: %#v", inputs)
		logger.Error("Error Inputs: %#v", in)
		logger.Error("Map Error: %s", err)
		return nil, err
	}
	o := FromPyObject(out)
	if out, ok := o.([]any); ok {
		return out, nil
	}
	return nil, fmt.Errorf("incorrect return type: %s", out)
}

func (p *PyCode) EvaluateBool(method string, inputs ...map[string]interface{}) (bool, error) {
	fun := p.module.Globals[method]
	in := py.Tuple{}
	for _, i := range inputs {
		data := PyObject(i)
		in = append(in, data)
	}
	out, err := py.Call(fun, in, nil)
	if err != nil {
		py.TracebackDump(err)
		logger.Error("Map Error: %s", err)
		return false, err
	}
	o := FromPyObject(out)
	if out, ok := o.(bool); ok {
		return out, nil
	}
	return false, fmt.Errorf("incorrect return type: %s", out)
}

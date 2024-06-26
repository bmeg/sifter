package transform

import (
	"fmt"
	"reflect"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/task"
)

var DefaultEngine = "python"
var PipeSize = 20

type MapProcessor interface {
	PoolReady() bool
	Process(map[string]any) map[string]any
}

type FlatMapProcessor interface {
	PoolReady() bool
	Process(map[string]any) []map[string]any
}

type StreamProcessor interface {
	//PoolReady() bool
	Process(chan map[string]any, chan map[string]any)
}

type ReduceProcessor interface {
	GetInit() map[string]any
	GetKey(map[string]any) string
	Reduce(key string, a map[string]any, b map[string]any) map[string]any
}

type NodeProcessor interface {
	Process(map[string]any) []map[string]any
}

type AccumulateProcessor interface {
	GetKey(map[string]any) string
	Accumulate(key string, value []map[string]any) map[string]any
}

type JoinProcessor interface {
	GetRightPipeline() string
	Process(left chan map[string]any, right chan map[string]any, out chan map[string]any)
}

type Processor interface {
	Close()
}

type Transform interface {
	Init(t task.RuntimeTask) (Processor, error)
}

type Step struct {
	Docs              string              `json:"docs"`
	From              *FromStep           `json:"from"`
	Split             *SplitStep          `json:"split"`
	FieldParse        *FieldParseStep     `json:"fieldParse" jsonschema_description:"fieldParse to run"`
	FieldType         *FieldTypeStep      `json:"fieldType" jsonschema_description:"Change type of a field (ie string -> integer)"`
	ObjectValidate    *ObjectValidateStep `json:"objectValidate" jsonschema_description:"Validate a JSON schema based object"`
	Emit              *EmitStep           `json:"emit" jsonschema_description:"Write to unstructured JSON file"`
	Filter            *FilterStep         `json:"filter"`
	Clean             *CleanStep          `json:"clean"`
	Debug             *DebugStep          `json:"debug" jsonschema_description:"Print message contents to stdout"`
	RegexReplace      *RegexReplaceStep   `json:"regexReplace"`
	Project           *ProjectStep        `json:"project" jsonschema_description:"Run a projection mapping message"`
	Map               *MapStep            `json:"map" jsonschema_description:"Apply a single function to all records"`
	Plugin            *PluginStep         `json:"plugin"`
	FlatMap           *FlatMapStep        `json:"flatmap" jsonschema_description:"Apply a single function to all records, flatten list responses"`
	Reduce            *ReduceStep         `json:"reduce"`
	Distinct          *DistinctStep       `json:"distinct"`
	FieldProcess      *FieldProcessStep   `json:"fieldProcess" jsonschema_description:"Take an array field from a message and run in child transform"`
	DropNull          *DropNullStep       `json:"dropNull"`
	Lookup            *LookupStep         `json:"lookup"`
	IntervalIntersect *IntervalStep       `json:"intervalIntersect"`
	Hash              *HashStep           `json:"hash"`
	GraphBuild        *GraphBuildStep     `json:"graphBuild"`
	TableWrite        *TableWriteStep     `json:"tableWrite"`
	Accumulate        *AccumulateStep     `json:"accumulate"`
	UUID              *UUIDStep           `json:"uuid"`
}

type Pipe []Step

func (ts Step) Init(t task.RuntimeTask) (Processor, error) {
	v := reflect.ValueOf(ts)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		x := f.Interface()
		if z, ok := x.(Transform); ok {
			if !f.IsNil() {
				return z.Init(t)
			}
		}
	}
	return nil, fmt.Errorf(("Transform not defined"))
}

func (ts Step) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	v := reflect.ValueOf(ts)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		x := f.Interface()
		if z, ok := x.(config.Configurable); ok {
			if !f.IsNil() {
				out = append(out, z.GetConfigFields()...)
			}
		}
	}
	return out
}

func (ts Step) GetEmitters() []string {
	if ts.Emit != nil {
		return []string{ts.Emit.Name}
	}
	if ts.GraphBuild != nil {
		return []string{"vertex", "edge"}
	}
	return []string{}
}

func (ts Step) GetOutputs() []string {
	if ts.TableWrite != nil {
		return []string{ts.TableWrite.Output}
	}
	return []string{}
}

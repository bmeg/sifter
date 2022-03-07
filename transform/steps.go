package transform

import (
	"fmt"
	"reflect"

	"github.com/bmeg/sifter/task"
)

var DefaultEngine = "python"
var PipeSize = 20

type Processor interface {
	Process(map[string]any) []map[string]any
	Close()
}

type Transform interface {
	Init(t task.RuntimeTask) (Processor, error)
}

type Step struct {
	FieldMap       *FieldMapStep       `json:"fieldMap" jsonschema_description:"fieldMap to run"`
	FieldType      *FieldTypeStep      `json:"fieldType" jsonschema_description:"Change type of a field (ie string -> integer)"`
	ObjectCreate   *ObjectCreateStep   `json:"objectCreate" jsonschema_description:"Create a JSON schema based object"`
	Emit           *EmitStep           `json:"emit" jsonschema_description:"Write to unstructured JSON file"`
	Filter         *FilterStep         `json:"filter"`
	Clean          *CleanStep          `json:"clean"`
	Debug          *DebugStep          `json:"debug" jsonschema_description:"Print message contents to stdout"`
	RegexReplace   *RegexReplaceStep   `json:"regexReplace"`
	AlleleID       *AlleleIDStep       `json:"alleleID" jsonschema_description:"Generate a standardized allele hash ID"`
	Project        *ProjectStep        `json:"project" jsonschema_description:"Run a projection mapping message"`
	Map            *MapStep            `json:"map" jsonschema_description:"Apply a single function to all records"`
	Reduce         *ReduceStep         `json:"reduce"`
	Distinct       *DistinctStep       `json:"distinct"`
	FieldProcess   *FieldProcessStep   `json:"fieldProcess" jsonschema_description:"Take an array field from a message and run in child transform"`
	TableReplace   *TableReplaceStep   `json:"tableReplace" jsonschema_description:"Load in TSV to map a fields values"`
	TableLookup    *TableLookupStep    `json:"tableLookup"`
	JSONFileLookup *JSONFileLookupStep `json:"jsonLookup"`
	GraphBuild     *GraphBuildStep     `json:"graphBuild"`
	//TableWrite     *TableWriteStep     `json:"tableWrite" jsonschema_description:"Write out a TSV"`
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

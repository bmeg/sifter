package transform

import (
	"log"
	"regexp"

	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

var DefaultEngine = "python"
var PipeSize = 20

type Process interface {
	Start(in chan map[string]interface{}) map[string]interface{}
	Close()
}

type Transform interface {
	Init(task.RuntimeTask) Process
}

type Step struct {
	FieldMap  *FieldMapStep  `json:"fieldMap" jsonschema_description:"fieldMap to run"`
	FieldType *FieldTypeStep `json:"fieldType" jsonschema_description:"Change type of a field (ie string -> integer)"`
	//ObjectCreate   *ObjectCreateStep   `json:"objectCreate" jsonschema_description:"Create a JSON schema based object"`
	Emit         *EmitStep         `json:"emit" jsonschema_description:"Write to unstructured JSON file"`
	Filter       *FilterStep       `json:"filter"`
	Clean        *CleanStep        `json:"clean"`
	Debug        *DebugStep        `json:"debug" jsonschema_description:"Print message contents to stdout"`
	RegexReplace *RegexReplaceStep `json:"regexReplace"`
	AlleleID     *AlleleIDStep     `json:"alleleID" jsonschema_description:"Generate a standardized allele hash ID"`
	Project      *ProjectStep      `json:"project" jsonschema_description:"Run a projection mapping message"`
	Map          *MapStep          `json:"map" jsonschema_description:"Apply a single function to all records"`
	Reduce       *ReduceStep       `json:"reduce"`
	Distinct     *DistinctStep     `json:"distinct"`
	FieldProcess *FieldProcessStep `json:"fieldProcess" jsonschema_description:"Take an array field from a message and run in child transform"`
	//TableWrite     *TableWriteStep     `json:"tableWrite" jsonschema_description:"Write out a TSV"`
	TableReplace   *TableReplaceStep   `json:"tableReplace" jsonschema_description:"Load in TSV to map a fields values"`
	TableLookup    *TableLookupStep    `json:"tableLookup"`
	JSONFileLookup *JSONFileLookupStep `json:"jsonLookup"`
}

type Pipe []Step

func contains(s []string, q string) bool {
	for _, i := range s {
		if i == q {
			return true
		}
	}
	return false
}

func (ts Step) Init(task task.RuntimeTask) error {
	log.Printf("Doing Step Init")
	if ts.Filter != nil {
		ts.Filter.Init(task)
	} else if ts.FieldProcess != nil {
		ts.FieldProcess.Init(task)
	} else if ts.RegexReplace != nil {
		re, _ := evaluate.ExpressionString(ts.RegexReplace.Regex, task.GetInputs(), nil)
		ts.RegexReplace.reg, _ = regexp.Compile(re)
	} else if ts.Map != nil {
		log.Printf("About to init map")
		ts.Map.Init(task)
	} else if ts.Reduce != nil {
		ts.Reduce.Init(task)
	} else if ts.Distinct != nil {
		ts.Distinct.Init(task)
		//} else if ts.TableWrite != nil {
		//	ts.TableWrite.Init(task)
	} else if ts.TableReplace != nil {
		err := ts.TableReplace.Init(task)
		if err != nil {
			log.Printf("TableReplace err: %s", err)
		}
		return err
	} else if ts.TableLookup != nil {
		err := ts.TableLookup.Init(task)
		if err != nil {
			log.Printf("TableLookup err: %s", err)
		}
		return err
	} else if ts.JSONFileLookup != nil {
		err := ts.JSONFileLookup.Init(task)
		if err != nil {
			log.Printf("JSONFileLookup err: %s", err)
		}
		return err
	}
	return nil
}

func (ts Step) Close() {
	if ts.Filter != nil {
		ts.Filter.Close()
	} else if ts.FieldProcess != nil {
		ts.FieldProcess.Close()
	} else if ts.Map != nil {
		ts.Map.Close()
	} else if ts.Distinct != nil {
		ts.Distinct.Close()
		//} else if ts.TableWrite != nil {
		//	ts.TableWrite.Close()
	}
}

func (ts Step) Start(in chan map[string]interface{},
	task task.RuntimeTask, wg *sync.WaitGroup) chan map[string]interface{} {

	out := make(chan map[string]interface{}, PipeSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		if ts.FieldMap != nil {
			for i := range in {
				out <- ts.FieldMap.Run(i, task)
			}
		} else if ts.FieldType != nil {
			for i := range in {
				out <- ts.FieldType.Run(i, task)
			}
		} else if ts.Filter != nil {
			fOut, _ := ts.Filter.Start(in, task, wg)
			for i := range fOut {
				out <- i
			}
		} else if ts.Clean != nil {
			fOut, _ := ts.Clean.Start(in, task, wg)
			for i := range fOut {
				out <- i
			}
		} else if ts.FieldProcess != nil {
			fOut, _ := ts.FieldProcess.Start(in, task, wg)
			for i := range fOut {
				out <- i
			}
		} else if ts.Debug != nil {
			for i := range in {
				out <- ts.Debug.Run(i, task)
			}
		} else if ts.RegexReplace != nil {
			for i := range in {
				out <- ts.RegexReplace.Run(i, task)
			}
		} else if ts.AlleleID != nil {
			for i := range in {
				out <- ts.AlleleID.Run(i, task)
			}
		} else if ts.Project != nil {
			for i := range in {
				out <- ts.Project.Run(i, task)
			}
		} else if ts.Map != nil {
			for i := range in {
				o := ts.Map.Run(i, task)
				out <- o
			}
		} else if ts.Reduce != nil {
			for i := range in {
				ts.Reduce.Add(i, task)
			}
			for o := range ts.Reduce.Run() {
				out <- o
			}
		} else if ts.Distinct != nil {
			fOut, _ := ts.Distinct.Start(in, task, wg)
			for i := range fOut {
				out <- i
			}
			//} else if ts.ObjectCreate != nil {
			//	for i := range in {
			//		out <- ts.ObjectCreate.Run(i, task)
			//	}
		} else if ts.Emit != nil {
			for i := range in {
				out <- ts.Emit.Run(i, task)
			}
			//} else if ts.TableWrite != nil {
			//	for i := range in {
			//		out <- ts.TableWrite.Run(i, task)
			//	}
		} else if ts.TableReplace != nil {
			for i := range in {
				out <- ts.TableReplace.Run(i, task)
			}
		} else if ts.TableLookup != nil {
			for i := range in {
				out <- ts.TableLookup.Run(i, task)
			}
		} else if ts.JSONFileLookup != nil {
			for i := range in {
				out <- ts.JSONFileLookup.Run(i, task)
			}
		} else {
			log.Printf("Unknown field step")
		}
	}()
	return out
}

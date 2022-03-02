package transform

import (
	"fmt"
	"log"
	"regexp"

	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

var DefaultEngine = "python"
var PipeSize = 20

type ColumnReplaceStep struct {
	Column  string `json:"col"`
	Pattern string `json:"pattern"`
	Replace string `json:"replace"`
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
	TableWrite     *TableWriteStep     `json:"tableWrite" jsonschema_description:"Write out a TSV"`
	TableReplace   *TableReplaceStep   `json:"tableReplace" jsonschema_description:"Load in TSV to map a fields values"`
	TableLookup    *TableLookupStep    `json:"tableLookup"`
	JSONFileLookup *JSONFileLookupStep `json:"jsonLookup"`
	Fork           *ForkStep           `json:"fork" jsonschema_description:"Take message stream and split into multiple child transforms"`
	Cache          *CacheStep          `json:"cache" jsonschema_description:"Sub a child transform pipeline, caching the results in a database"`
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
	} else if ts.TableWrite != nil {
		ts.TableWrite.Init(task)
	} else if ts.TableReplace != nil {
		err := ts.TableReplace.Init(task)
		if err != nil {
			log.Printf("TableReplace err: %s", err)
		}
		return err
	} else if ts.Cache != nil {
		return ts.Cache.Init(task)
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
	} else if ts.Fork != nil {
		err := ts.Fork.Init(task)
		if err != nil {
			log.Printf("Fork err: %s", err)
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
	} else if ts.TableWrite != nil {
		ts.TableWrite.Close()
	} else if ts.Cache != nil {
		ts.Cache.Close()
	} else if ts.Fork != nil {
		ts.Fork.Close()
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
		} else if ts.ObjectCreate != nil {
			for i := range in {
				out <- ts.ObjectCreate.Run(i, task)
			}
		} else if ts.Emit != nil {
			for i := range in {
				out <- ts.Emit.Run(i, task)
			}
		} else if ts.TableWrite != nil {
			for i := range in {
				out <- ts.TableWrite.Run(i, task)
			}
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
		} else if ts.Cache != nil {
			outCache, err := ts.Cache.Start(in, task, wg)
			if err != nil {
				log.Printf("Cache err: %s", err)
			}
			for i := range outCache {
				out <- i
			}
		} else if ts.Fork != nil {
			outFork, err := ts.Fork.Start(in, task, wg)
			if err != nil {
				log.Printf("Fork err: %s", err)
			}
			for i := range outFork {
				out <- i
			}
		} else {
			log.Printf("Unknown field step")
		}
	}()
	return out
}

func (tp Pipe) Init(task task.RuntimeTask) error {
	log.Printf("Transform Pipe Init")
	for _, s := range tp {
		if err := s.Init(task); err != nil {
			return err
		}
	}
	return nil
}

func (tp Pipe) Close() {
	for _, s := range tp {
		s.Close()
	}
}

func (tp Pipe) Start(in chan map[string]interface{},
	task task.RuntimeTask,
	wg *sync.WaitGroup) (chan map[string]interface{}, error) {

	log.Printf("Starting Transform Pipe")
	out := make(chan map[string]interface{}, 10)
	wg.Add(1)
	go func() {
		defer close(out)
		cwg := &sync.WaitGroup{}
		//connect the input stream to the processing chain
		cur := in
		for i, s := range tp {
			cur = s.Start(cur, task.Child(fmt.Sprintf("%d", i)), cwg)
		}
		for i := range cur {
			out <- i
		}
		cwg.Wait()
		wg.Done()
		log.Printf("Ending Transform Pipe")
	}()
	return out, nil
}

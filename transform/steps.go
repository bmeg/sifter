package transform

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"crypto/sha1"
	"encoding/json"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/pipeline"
)

var DefaultEngine = "python"
var PipeSize = 20

type ColumnReplaceStep struct {
	Column  string `json:"col"`
	Pattern string `json:"pattern"`
	Replace string `json:"replace"`
}

type FieldMapStep struct {
	Column string `json:"col"`
	Sep    string `json:"sep"`
}

type FieldTypeStep struct {
	Column string `json:"col"`
	Type   string `json:"type"`
}

type RegexReplaceStep struct {
	Column  string `json:"col"`
	Regex   string `json:"regex"`
	Replace string `json:"replace"`
	Dest    string `json:"dst"`
	reg     *regexp.Regexp
}

type AlleleIDStep struct {
	Prefix         string `json:"prefix"`
	Genome         string `json:"genome"`
	Chromosome     string `json:"chromosome"`
	Start          string `json:"start"`
	End            string `json:"end"`
	ReferenceBases string `json:"reference_bases"`
	AlternateBases string `json:"alternate_bases"`
	Dest           string `json:"dst"`
}

type DebugStep struct {
	Label string `json:"label"`
}

type CacheStep struct {
	Transform Pipe `json:"transform"`
}

type Step struct {
	FieldMap     *FieldMapStep     `json:"fieldMap" jsonschema_description:"fieldMap to run"`
	FieldType    *FieldTypeStep    `json:"fieldType" jsonschema_description:"Change type of a field (ie string -> integer)"`
	ObjectCreate *ObjectCreateStep `json:"objectCreate" jsonschema_description:"Create a JSON schema based object"`
	Emit         *EmitStep         `json:"emit" jsonschema_description:"Write to unstructured JSON file"`
	Filter       *FilterStep       `json:"filter"`
	Clean        *CleanStep        `json:"clean"`
	Debug        *DebugStep        `json:"debug" jsonschema_description:"Print message contents to stdout"`
	RegexReplace *RegexReplaceStep `json:"regexReplace"`
	AlleleID     *AlleleIDStep     `json:"alleleID" jsonschema_description:"Generate a standardized allele hash ID"`
	Project      *ProjectStep      `json:"project" jsonschema_description:"Run a projection mapping message"`
	Map          *MapStep          `json:"map" jsonschema_description:"Apply a single function to all records"`
	Reduce       *ReduceStep       `json:"reduce"`
	FieldProcess *FieldProcessStep `json:"fieldProcess" jsonschema_description:"Take an array field from a message and run in child transform"`
	TableWrite   *TableWriteStep   `json:"tableWrite" jsonschema_description:"Write out a TSV"`
	TableReplace *TableReplaceStep `json:"tableReplace" jsonschema_description:"Load in TSV to map a fields values"`
	TableLookup *TableLookupStep `json:"tableLookup"`
	JSONFileLookup *JSONFileLookupStep `json:"jsonLookup"`
	Fork         *ForkStep         `json:"fork" jsonschema_description:"Take message stream and split into multiple child transforms"`
	Cache        *CacheStep        `json:"cache" jsonschema_description:"Sub a child transform pipeline, caching the results in a database"`
}

type Pipe []Step

func (fm FieldMapStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	o := map[string]interface{}{}
	for x, y := range i {
		o[x] = y
	}
	if v, ok := i[fm.Column]; ok {
		if vStr, ok := v.(string); ok {
			a := strings.Split(vStr, fm.Sep)
			t := map[string]interface{}{}
			for _, s := range a {
				kv := strings.Split(s, "=")
				if len(kv) > 1 {
					t[kv[0]] = kv[1]
				} else {
					t[kv[0]] = true
				}
			}
			o[fm.Column] = t
		}
	}
	return o
}

func (fs FieldTypeStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	o := map[string]interface{}{}
	for x, y := range i {
		o[x] = y
	}
	if fs.Type == "int" {
		if val, ok := i[fs.Column]; ok {
			if vStr, ok := val.(string); ok {
				if d, err := strconv.ParseInt(vStr, 10, 64); err == nil {
					o[fs.Column] = d
				}
			}
		}
	}
	return o
}

func contains(s []string, q string) bool {
	for _, i := range s {
		if i == q {
			return true
		}
	}
	return false
}

func (re RegexReplaceStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	col, _ := evaluate.ExpressionString(re.Column, task.Inputs, i)
	replace, _ := evaluate.ExpressionString(re.Replace, task.Inputs, i)
	dst, _ := evaluate.ExpressionString(re.Dest, task.Inputs, i)

	o := re.reg.ReplaceAllString(col, replace)
	z := map[string]interface{}{}
	for x, y := range i {
		z[x] = y
	}
	z[dst] = o
	return z
}

func (al AlleleIDStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {

	genome, _ := evaluate.ExpressionString(al.Genome, task.Inputs, i)
	chromosome, _ := evaluate.ExpressionString(al.Chromosome, task.Inputs, i)
	start, _ := evaluate.ExpressionString(al.Start, task.Inputs, i)
	end, _ := evaluate.ExpressionString(al.End, task.Inputs, i)
	ref, _ := evaluate.ExpressionString(al.ReferenceBases, task.Inputs, i)
	alt, _ := evaluate.ExpressionString(al.AlternateBases, task.Inputs, i)

	id := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		genome, chromosome,
		start, end,
		ref,
		alt)
	//log.Printf("AlleleStr: %s", id)
	idSha1 := fmt.Sprintf("%x", sha1.Sum([]byte(id)))
	if al.Prefix != "" {
		idSha1 = al.Prefix + idSha1
	}
	o := map[string]interface{}{}
	for k, v := range i {
		o[k] = v
	}
	o[al.Dest] = idSha1
	return o
}

func (db DebugStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	s, _ := json.Marshal(i)
	log.Printf("DebugData %s: %s", db.Label, s)
	return i
}

func (ts Step) Init(task *pipeline.Task) error {
	log.Printf("Doing Step Init")
	if ts.Filter != nil {
		ts.Filter.Init(task)
	} else if ts.FieldProcess != nil {
		ts.FieldProcess.Init(task)
	} else if ts.RegexReplace != nil {
		re, _ := evaluate.ExpressionString(ts.RegexReplace.Regex, task.Inputs, nil)
		ts.RegexReplace.reg, _ = regexp.Compile(re)
	} else if ts.Map != nil {
		log.Printf("About to init map")
		ts.Map.Init(task)
	} else if ts.Reduce != nil {
		ts.Reduce.Init(task)
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
	} else if ts.TableWrite != nil {
		ts.TableWrite.Close()
	} else if ts.Cache != nil {
		ts.Cache.Close()
	} else if ts.Fork != nil {
		ts.Fork.Close()
	}
}

func (ts Step) Start(in chan map[string]interface{},
	task *pipeline.Task, wg *sync.WaitGroup) chan map[string]interface{} {

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

func (tp Pipe) Init(task *pipeline.Task) error {
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
	task *pipeline.Task,
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

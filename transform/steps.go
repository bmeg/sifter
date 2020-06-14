package transform

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/bmeg/sifter/emitter"
	"github.com/bmeg/sifter/evaluate"

	"encoding/json"

	"crypto/sha1"
	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/pipeline"

	"github.com/cnf/structhash"
)

var DEFAULT_ENGINE = "python"

type ObjectCreateStep struct {
	Class string `json:"class" jsonschema_description:"Object class, should match declared class in JSON Schema"`
	Name  string `json:"name" jsonschema_description:"domain name of stream, to seperate it from other output streams of the same output type"`
}

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

type FilterStep struct {
	Column string        `json:"col"`
	Match  string        `json:"match"`
	Exists bool          `json:"exists"`
	Method string        `json:"method"`
	Python string        `json:"python"`
	Steps  TransformPipe `json:"steps"`
	inChan chan map[string]interface{}
	proc   evaluate.Processor
}

type RegexReplaceStep struct {
	Column  string `json:"col"`
	Regex   string `json:"regex"`
	Replace string `json:"replace"`
	Dest    string `json:"dst"`
	reg     *regexp.Regexp
}

type AlleleIDStep struct {
	Prefix         string `json:prefix`
	Genome         string `json:"genome"`
	Chromosome     string `json:"chromosome"`
	Start          string `json:"start"`
	End            string `json:"end"`
	ReferenceBases string `json:"reference_bases"`
	AlternateBases string `json:"alternate_bases"`
	Dest           string `json:"dst"`
}

type ProjectStep struct {
	Mapping map[string]interface{} `json:"mapping" jsonschema_description:"New fields to be generated from template"`
}

type FieldProcessStep struct {
	Column  string            `json:"col"`
	Steps   TransformPipe     `json:"steps"`
	Mapping map[string]string `json:"mapping"`
	inChan  chan map[string]interface{}
}

type TableWriteStep struct {
	Output  string   `json:"output" jsonschema_description:"Name of file to create"`
	Columns []string `json:"columns" jsonschema_description:"Columns to be written into table file"`
	emit    emitter.TableEmitter
}

type TableReplaceStep struct {
	Input string `json:"input"`
	Field string `json:"field"`
	table map[string]string
}

type TableProjectStep struct {
	Input string `json:"input"`
	table map[string]string
}

type DebugStep struct {
	Label string `json:"label"`
}

type CacheStep struct {
	Transform TransformPipe `json:"transform"`
}

type TransformStep struct {
	FieldMap     *FieldMapStep     `json:"fieldMap" jsonschema_description:"fieldMap to run"`
	FieldType    *FieldTypeStep    `json:"fieldType" jsonschema_description:"Change type of a field (ie string -> integer)"`
	ObjectCreate *ObjectCreateStep `json:"objectCreate" jsonschema_description:"Create a JSON schema based object"`
	Filter       *FilterStep       `json:"filter"`
	Debug        *DebugStep        `json:"debug" jsonschema_description:"Print message contents to stdout"`
	RegexReplace *RegexReplaceStep `json:"regexReplace"`
	AlleleID     *AlleleIDStep     `json:"alleleID" jsonschema_description:"Generate a standardized allele hash ID"`
	Project      *ProjectStep      `json:"project" jsonschema_description:"Run a projection mapping message"`
	Map          *MapStep          `json:"map" jsonschema_description:"Apply a single function to all records"`
	Reduce       *ReduceStep       `json:"reduce"`
	FieldProcess *FieldProcessStep `json:"fieldProcess" jsonschema_description:"Take an array field from a message and run in child transform"`
	TableWrite   *TableWriteStep   `json:"tableWrite" jsonschema_description:"Write out a TSV"`
	TableReplace *TableReplaceStep `json:"tableReplace" jsonschema_description:"Load in TSV to map a fields values"`
	TableProject *TableProjectStep `json:"tableProject"`
	Fork         *ForkStep         `json:"fork" jsonschema_description:"Take message stream and split into multiple child transforms"`
	Cache        *CacheStep        `json:"cache" jsonschema_description:"Sub a child transform pipeline, caching the results in a database"`
}

type TransformPipe []TransformStep

type ForkStep struct {
	Transform []TransformPipe `json:"transform"`
	pipes     []chan map[string]interface{}
}

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

func (ts ObjectCreateStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	task.Runtime.EmitObject(ts.Name, ts.Class, i)
	return i
}

func (tw *TableWriteStep) Init(task *pipeline.Task) {
	tw.emit = task.Runtime.EmitTable(tw.Output, tw.Columns)
}

func (tw *TableWriteStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	if err := tw.emit.EmitRow(i); err != nil {
		log.Printf("Row Error: %s", err)
	}
	return i
}

func (tw *TableWriteStep) Close() {
	tw.emit.Close()
}

func (tr *TableReplaceStep) Init(task *pipeline.Task) error {
	input, err := evaluate.ExpressionString(tr.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)
	if err != nil {
		return err
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)

	inputStream, err := golib.ReadFileLines(inputPath)
	if err != nil {
		return err
	}
	tr.table = map[string]string{}
	for line := range inputStream {
		if len(line) > 0 {
			row := strings.Split(string(line), "\t")
			tr.table[row[0]] = row[1]
		}
	}
	return nil
}

func (tw *TableReplaceStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {

	if _, ok := i[tw.Field]; ok {
		out := map[string]interface{}{}
		for k, v := range i {
			if k == tw.Field {
				if x, ok := v.(string); ok {
					if n, ok := tw.table[x]; ok {
						out[k] = n
					} else {
						out[k] = x
					}
				} else if x, ok := v.([]interface{}); ok {
					o := []interface{}{}
					for _, y := range x {
						if z, ok := y.(string); ok {
							if n, ok := tw.table[z]; ok {
								o = append(o, n)
							} else {
								o = append(o, z)
							}
						}
					}
					out[k] = o
				} else {
					out[k] = v
				}
			} else {
				out[k] = v
			}
		}
		return out
	}
	return i
}

func (tr *TableProjectStep) Init(task *pipeline.Task) error {
	input, err := evaluate.ExpressionString(tr.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)
	if err != nil {
		return err
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading Translation file: %s", inputPath)

	inputStream, err := golib.ReadFileLines(inputPath)
	if err != nil {
		return err
	}
	tr.table = map[string]string{}
	for line := range inputStream {
		if len(line) > 0 {
			row := strings.Split(string(line), "\t")
			tr.table[row[0]] = row[1]
		}
	}
	return nil
}

func (tw *TableProjectStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {

	out := map[string]interface{}{}
	for k, v := range i {
		if n, ok := tw.table[k]; ok {
			out[n] = v
		} else {
			out[k] = v
		}
	}
	return out
}

func (fs *FilterStep) Init(task *pipeline.Task) {
	if fs.Python != "" && fs.Method != "" {
		log.Printf("Starting Map: %s", fs.Python)
		e := evaluate.GetEngine(DEFAULT_ENGINE)
		c, err := e.Compile(fs.Python, fs.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		fs.proc = c
	}
	fs.Steps.Init(task)
}

func (fs FilterStep) Start(in chan map[string]interface{}, task *pipeline.Task, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	out := make(chan map[string]interface{}, 10)
	fs.inChan = make(chan map[string]interface{}, 100)
	tout, _ := fs.Steps.Start(fs.inChan, task.Child("filter"), wg)
	go func() {
		//Filter does not emit the output of its sub pipeline, but it has to digest it
		for range tout {
		}
	}()

	go func() {
		//Filter emits a copy of its input, without changing it
		defer close(out)
		defer close(fs.inChan)
		for i := range in {
			fs.run(i, task)
			out <- i
		}
	}()
	return out, nil
}

func (fs FilterStep) run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	if fs.Python != "" && fs.Method != "" {
		out, err := fs.proc.EvaluateBool(i)
		if err != nil {
			log.Printf("Filter Error: %s", err)
		}
		if out {
			fs.inChan <- i
		}
		return i
	}
	col, err := evaluate.ExpressionString(fs.Column, task.Inputs, i)
	if fs.Exists {
		if err != nil {
			return i
		}
	}
	match, _ := evaluate.ExpressionString(fs.Match, task.Inputs, i)
	if col == match {
		fs.inChan <- i
	}
	return i
}

func (fs FilterStep) Close() {
	if fs.proc != nil {
		fs.proc.Close()
	}
	fs.Steps.Close()
}

func (fs *FieldProcessStep) Init(task *pipeline.Task) {
	fs.inChan = make(chan map[string]interface{}, 100)
	//fs.Steps.Start(fs.inChan, task, wg)
}

func (fs FieldProcessStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	if v, ok := i[fs.Column]; ok {
		if vList, ok := v.([]interface{}); ok {
			for _, l := range vList {
				if m, ok := l.(map[string]interface{}); ok {
					r := map[string]interface{}{}
					for k, v := range m {
						r[k] = v
					}
					for k, v := range fs.Mapping {
						val, _ := evaluate.ExpressionString(v, task.Inputs, i)
						r[k] = val
					}
					fs.inChan <- r
				} else {
					log.Printf("Incorrect Field Type: %s", l)
				}
			}
		} else {
			log.Printf("Field list incorrect type: %s", v)
		}
	} else {
		log.Printf("Field %s missing", fs.Column)
	}
	return i
}

func (fs FieldProcessStep) Close() {
	close(fs.inChan)
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

func valueRender(v interface{}, task *pipeline.Task, row map[string]interface{}) (interface{}, error) {
	if vStr, ok := v.(string); ok {
		return evaluate.ExpressionString(vStr, task.Inputs, row)
	} else if vMap, ok := v.(map[string]interface{}); ok {
		o := map[string]interface{}{}
		for key, val := range vMap {
			o[key], _ = valueRender(val, task, row)
		}
		return o, nil
	} else if vArray, ok := v.([]interface{}); ok {
		o := []interface{}{}
		for _, val := range vArray {
			j, _ := valueRender(val, task, row)
			o = append(o, j)
		}
		return o, nil
	} else if vArray, ok := v.([]string); ok {
		o := []string{}
		for _, vStr := range vArray {
			j, _ := evaluate.ExpressionString(vStr, task.Inputs, row)
			o = append(o, j)
		}
		return o, nil
	}
	return v, nil
}

func (pr ProjectStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {

	o := map[string]interface{}{}
	for k, v := range i {
		o[k] = v
	}

	for k, v := range pr.Mapping {
		o[k], _ = valueRender(v, task, i)
	}
	return o
}

func (db DebugStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	s, _ := json.Marshal(i)
	log.Printf("DebugData %s: %s", db.Label, s)
	return i
}

func (cs *CacheStep) Init(task *pipeline.Task) error {
	return cs.Transform.Init(task)
}

func (cs *CacheStep) Start(in chan map[string]interface{}, task *pipeline.Task, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	log.Printf("Starting Cache: %s", task.Name)

	ds, err := task.GetDataStore()
	if err != nil {
		log.Printf("Cache setup error: %s", err)
	}

	if ds == nil {
		log.Printf("No cache setup")
		out, err := cs.Transform.Start(in, task, wg)
		return out, err
	}

	out := make(chan map[string]interface{}, 10)
	go func() {
		defer close(out)
		for i := range in {
			hash, err := structhash.Hash(i, 1)
			if err == nil {
				key := fmt.Sprintf("%s.%s", task.Name, hash)
				log.Printf("Cache Key: %s.%s", task.Name, hash)
				if ds.HasRecordStream(key) {
					log.Printf("Cache Hit")
					for j := range ds.GetRecordStream(key) {
						out <- j
					}
				} else {
					log.Printf("Cache Miss")

					manIn := make(chan map[string]interface{}, 10)
					manIn <- i
					close(manIn)

					cacheIn := make(chan map[string]interface{}, 10)
					go ds.SetRecordStream(key, cacheIn)

					newWG := &sync.WaitGroup{}

					tOut, _ := cs.Transform.Start(manIn, task, newWG)
					for j := range tOut {
						log.Printf("Cache Calc out: %s", j)
						cacheIn <- j
						out <- j
					}
					close(cacheIn)
				}
			} else {
				log.Printf("Hashing Error")
			}
		}
	}()

	return out, nil
}

func (cs *CacheStep) Close() {

}

func (fs *ForkStep) Init(task *pipeline.Task) error {
	fs.pipes = []chan map[string]interface{}{}
	//for _, t := range fs.Transform {
	//  p := make(chan map[string]interface{}, 100)
	//t.Start(p, task, wg)
	//  fs.pipes = append(fs.pipes, p)
	//}
	return nil
}

func (fs *ForkStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	for _, p := range fs.pipes {
		p <- i
	}
	return i
}

func (fs *ForkStep) Close() {
	for _, p := range fs.pipes {
		close(p)
	}
}

func (ts TransformStep) Init(task *pipeline.Task) error {
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
	} else if ts.TableProject != nil {
		err := ts.TableProject.Init(task)
		if err != nil {
			log.Printf("TableProject err: %s", err)
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

func (ts TransformStep) Close() {
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

func (ts TransformStep) Start(in chan map[string]interface{},
	task *pipeline.Task, wg *sync.WaitGroup) chan map[string]interface{} {

	out := make(chan map[string]interface{}, 100)
	go func() {
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
		} else if ts.FieldProcess != nil {
			for i := range in {
				out <- ts.FieldProcess.Run(i, task)
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
		} else if ts.TableWrite != nil {
			for i := range in {
				out <- ts.TableWrite.Run(i, task)
			}
		} else if ts.TableReplace != nil {
			for i := range in {
				out <- ts.TableReplace.Run(i, task)
			}
		} else if ts.TableProject != nil {
			for i := range in {
				out <- ts.TableProject.Run(i, task)
			}
		} else if ts.Cache != nil {
			outCache, err := ts.Cache.Start(in, task, wg)
			if err != nil {
				log.Printf("TableProject err: %s", err)
			}
			for i := range outCache {
				out <- i
			}
		} else if ts.Fork != nil {
			for i := range in {
				out <- ts.Fork.Run(i, task)
			}
		} else {
			log.Printf("Unknown field step")
		}
	}()
	return out
}

func (tp TransformPipe) Init(task *pipeline.Task) error {
	log.Printf("TransformPipe Init")
	for _, s := range tp {
		if err := s.Init(task); err != nil {
			return err
		}
	}
	return nil
}

func (tp TransformPipe) Close() {
	for _, s := range tp {
		s.Close()
	}
}

func (tp TransformPipe) Start(in chan map[string]interface{},
	task *pipeline.Task,
	wg *sync.WaitGroup) (chan map[string]interface{}, error) {

	log.Printf("Starting TransformPipe")
	out := make(chan map[string]interface{}, 10)
	wg.Add(1)
	go func() {
		defer close(out)
		//connect the input stream to the processing chain
		cur := in
		for i, s := range tp {
			cur = s.Start(cur, task.Child(fmt.Sprintf("%d", i)), wg)
		}
		for i := range cur {
			out <- i
		}
		wg.Done()
		log.Printf("Ending TransformPipe")
	}()
	return out, nil
}


package manager


import (
  "os"
  "io"
  "log"
  "fmt"
  "strings"
  "sync"
  "regexp"
  "strconv"

  "encoding/json"

  "crypto/sha1"
  "compress/gzip"

  "encoding/csv"
  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/grip/gripql"
  "github.com/bmeg/grip/protoutil"
)

type EdgeCreateStep struct {
  Gid   string `json:"gid"`
	To    string `json:"to"`
	From  string `json:"from"`
	Label string `json:"label"`
  Exclude []string `json:"exclude"`
  Include []string `json:"include"`
}

type VertexCreateStep struct {
	Gid   string `json:"gid"`
	Label string `json:"label"`
  Exclude []string `json:"exclude"`
  Include []string `json:"include"`
}

type ObjectCreateStep struct {
  Class  string `json:"class"`
}

type ColumnReplaceStep struct {
	Column  string `json:"col"`
	Pattern string `json:"pattern"`
	Replace string `json:"replace"`
}

type FieldMapStep struct {
  Column  string `json:"col"`
  Sep     string `json:"sep"`
}

type FieldTypeStep struct {
  Column  string `json:"col"`
  Type    string `json:"type"`
}

type FilterStep struct {
  Column  string `json:"col"`
  Match   string `json:"match"`
  Steps   TransformPipe `json:"steps"`
  inChan  chan map[string]interface{}
}

type RegexReplaceStep struct {
  Column string `json:"col"`
  Regex  string `json:"regex"`
  Replace string `json:"replace"`
  Dest    string `json:"dst"`
  reg     *regexp.Regexp
}


type AlleleIDStep struct {
  Prefix   string      `json:prefix`
  Genome   string      `json:"genome"`
  Chromosome string    `json:"chromosome"`
  Start    string      `json:"start"`
  End      string      `json:"end"`
  ReferenceBases string `json:"reference_bases"`
  AlternateBases string `json:"alternate_bases"`
  Dest     string       `json:"dst"`
}

type ProjectStep struct {
  Mapping map[string]string `json:"mapping"`
}

type FieldProcessStep struct {
  Column string                        `json:"col"`
  Steps   TransformPipe                `json:"steps"`
  Mapping map[string]string            `json:"mapping"`
  inChan  chan map[string]interface{}
}

type DebugStep struct {}

type TransformStep struct {
  FieldMap      *FieldMapStep          `json:"fieldMap"`
  FieldType     *FieldTypeStep         `json:"fieldType"`
  EdgeCreate    *EdgeCreateStep        `json:"edgeCreate"`
  VertexCreate  *VertexCreateStep      `json:"vertexCreate"`
  ObjectCreate  *ObjectCreateStep      `json:"objectCreate"`
  Filter        *FilterStep            `json:"filter"`
  Debug         *DebugStep             `json:"debug"`
  RegexReplace  *RegexReplaceStep      `json:"regexReplace"`
  AlleleID      *AlleleIDStep          `json:"alleleID"`
  Project       *ProjectStep           `json:"project"`
  Map           *MapStep               `json:"map"`
  Reduce        *ReduceStep            `json:"reduce"`
  FieldProcess  *FieldProcessStep      `json:"fieldProcess"`
  TableWrite    *TableWriteStep        `json:"tableWrite"`
}

type TransformPipe []TransformStep

type TableLoadStep struct {
  Input         string                 `json:"input"`
	RowSkip       int                    `json:"rowSkip"`
  SkipIfMissing bool                   `json:"skipIfMissing"`
  Columns       []string               `json:"columns"`
  Transform     []TransformPipe        `json:"transform"`
}

type TableWriteStep struct {
  Output       string   `json:"output"`
  Columns      []string `json:"columns"`
  out          *os.File
  writer       *csv.Writer
}

func (fm FieldMapStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  o := map[string]interface{}{}
  for x,y := range i {
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

func (fs FieldTypeStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  o := map[string]interface{}{}
  for x,y := range i {
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

func (ts VertexCreateStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  v := gripql.Vertex{}

  gid, err := evaluate.ExpressionString(ts.Gid, task.Inputs, i)
  if err != nil {
    log.Printf("Error: %s", err)
  }
  label, _ := evaluate.ExpressionString(ts.Label, task.Inputs, i)

  v.Gid = gid
  v.Label = label
  if ts.Exclude != nil && len(ts.Exclude) > 0 {
      t := map[string]interface{}{}
      for x,y := range i {
        if !contains(ts.Exclude, x) {
          t[x] = y
        }
      }
      v.Data = protoutil.AsStruct(t)
  } else if ts.Include != nil {
    t := map[string]interface{}{}
    for x,y := range i {
      if contains(ts.Exclude, x) {
        t[x] = y
      }
    }
    v.Data = protoutil.AsStruct(t)
  } else {
    v.Data = protoutil.AsStruct(i)
  }
  task.EmitVertex( &v )
  return i
}


func (ts EdgeCreateStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  e := gripql.Edge{}

  if ts.Gid != "" {
    gid, _ := evaluate.ExpressionString(ts.Gid, task.Inputs, i)
    e.Gid = gid
  }
  label, _ := evaluate.ExpressionString(ts.Label, task.Inputs, i)
  to, _ := evaluate.ExpressionString(ts.To, task.Inputs, i)
  from, _ := evaluate.ExpressionString(ts.From, task.Inputs, i)

  e.Label = label
  e.To = to
  e.From = from

  if ts.Exclude != nil && len(ts.Exclude) > 0 {
      t := map[string]interface{}{}
      for x,y := range i {
        if !contains(ts.Exclude, x) {
          t[x] = y
        }
      }
      e.Data = protoutil.AsStruct(t)
  } else if ts.Include != nil {
    t := map[string]interface{}{}
    for x,y := range i {
      if contains(ts.Exclude, x) {
        t[x] = y
      }
    }
    e.Data = protoutil.AsStruct(t)
  } else {
    e.Data = protoutil.AsStruct(i)
  }
  task.EmitEdge( &e )
  return i
}

func (ts ObjectCreateStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  //log.Printf("Create Object: %s", ts.Class)
  if task.Runtime.Schemas == nil {
    log.Printf("Schema not loaded")
    return i
  }
  if o, err := task.Runtime.Schemas.Generate(ts.Class, i); err == nil {
    for _, j := range o {
      if j.Vertex != nil {
        task.EmitVertex(j.Vertex)
      } else if j.Edge != nil {
        //log.Printf("Emitting: %s", j.Edge)
        task.EmitEdge(j.Edge)
      }
    }
  } else {
    s, _ := json.Marshal(i)
    log.Printf("Object Create Error: '%s' using '%s'", err, s)
  }
  return i
}


func (tw *TableWriteStep) Start(task *Task, wg *sync.WaitGroup) {
  path, _ := task.Path(tw.Output)
  tw.out, _ = os.Create(path)
  tw.writer = csv.NewWriter(tw.out)
  tw.writer.Comma = '\t'
  tw.writer.Write(tw.Columns)
}

func (tw *TableWriteStep)  Run(i map[string]interface{}, task *Task) map[string]interface{} {
  o := make([]string, len(tw.Columns))
  for j, k := range tw.Columns {
    if v, ok := i[k]; ok {
      if vStr, ok := v.(string); ok {
        o[j] = vStr
      }
    }
  }
  tw.writer.Write(o)
  return i
}

func (tw *TableWriteStep) Close() {
  tw.writer.Flush()
  tw.out.Close()
}

func (fs *FilterStep) Start(task *Task, wg *sync.WaitGroup) {
  fs.inChan = make(chan map[string]interface{}, 100)
  fs.Steps.Start(fs.inChan, task, wg)
}

func (fs FilterStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  col, _ := evaluate.ExpressionString(fs.Column, task.Inputs, i)
  match, _ := evaluate.ExpressionString(fs.Match, task.Inputs, i)
  if col == match {
    fs.inChan <- i
  }
  return i
}

func (fs FilterStep) Close() {
  close(fs.inChan)
}


func (fs *FieldProcessStep) Start(task *Task, wg *sync.WaitGroup) {
  fs.inChan = make(chan map[string]interface{}, 100)
  fs.Steps.Start(fs.inChan, task, wg)
}


func (fs FieldProcessStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
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
          }
        }
      }
  }
  return i
}

func (fs FieldProcessStep) Close() {
  close(fs.inChan)
}

func (re RegexReplaceStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  col, _ := evaluate.ExpressionString(re.Column, task.Inputs, i)
  replace, _ := evaluate.ExpressionString(re.Replace, task.Inputs, i)
  dst, _ := evaluate.ExpressionString(re.Dest, task.Inputs, i)

  o := re.reg.ReplaceAllString(col, replace)
  z := map[string]interface{}{}
  for x,y := range i {
    z[x] = y
  }
  z[dst] = o
  return z
}

func (al AlleleIDStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {

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

func (pr ProjectStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {

  o := map[string]interface{}{}
  for k, v := range i {
    o[k] = v
  }

  for k, v := range pr.Mapping {
    o[k], _ = evaluate.ExpressionString(v, task.Inputs, i)
  }
  return o
}

func (db DebugStep) Run(i map[string]interface{}, task *Task) map[string]interface{} {
  s, _ := json.Marshal(i)
  log.Printf("DebugData: %s", s)
  return i
}

func (ts TransformStep) Start(in chan map[string]interface{},
  task *Task, wg *sync.WaitGroup) chan map[string]interface{} {

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
    } else if ts.VertexCreate != nil {
      for i := range in {
        out <- ts.VertexCreate.Run(i, task)
      }
    } else if ts.EdgeCreate != nil {
      for i := range in {
        out <- ts.EdgeCreate.Run(i, task)
      }
    } else if ts.Filter != nil {
      ts.Filter.Start(task, wg)
      for i := range in {
        out <- ts.Filter.Run(i, task)
      }
      ts.Filter.Close()
    } else if ts.FieldProcess != nil {
      ts.FieldProcess.Start(task, wg)
      for i := range in {
        out <- ts.FieldProcess.Run(i, task)
      }
      ts.FieldProcess.Close()
    } else if ts.Debug != nil {
      for i := range in {
        out <- ts.Debug.Run(i, task)
      }
    } else if ts.RegexReplace != nil {
      re, _ := evaluate.ExpressionString(ts.RegexReplace.Regex, task.Inputs, nil)
      ts.RegexReplace.reg, _ = regexp.Compile(re)
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
      ts.Map.Start(task, wg)
      for i := range in {
        out <- ts.Map.Run(i, task)
      }
    } else if ts.Reduce != nil {
      ts.Reduce.Start(task, wg)
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
      ts.TableWrite.Start(task, wg)
      for i := range in {
        out <- ts.TableWrite.Run(i, task)
      }
      ts.TableWrite.Close()
    } else {
      log.Printf("Unknown field step")
    }
  }()
  return out
}

func (tp TransformPipe) Start( in chan map[string]interface{},
  task *Task,
  wg *sync.WaitGroup) {

    wg.Add(1)
    //connect the input stream to the processing chain
    cur := in
    for _, s := range tp {
      cur = s.Start(cur, task, wg)
    }

    go func () {
      //read the output pipe to pull all the data through the pipe
      for range cur {}
      wg.Done()
    }()
}


func (ml *TableLoadStep) Run(task *Task) error {
  log.Printf("Starting Table Load")
	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		if ml.SkipIfMissing {
			return nil
		}
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)
	fhd, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer fhd.Close()

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return err
		}
	} else {
    hd = fhd
  }

  r := csv.NewReader(hd)
  r.Comma = '\t'
  r.Comment = '#'

  var columns []string
  if ml.Columns != nil {
    columns = ml.Columns
  }

  procChan := []chan map[string]interface{}{}
  wg := &sync.WaitGroup{}
  for _, trans := range ml.Transform {
    i := make(chan map[string]interface{}, 100)
    trans.Start(i, task, wg)
    procChan = append(procChan, i)
  }
  rowSkip := ml.RowSkip

  for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
      log.Printf("Error %s", err)
			break
		}
    if rowSkip > 0 {
      rowSkip--
    } else {
      if columns == nil {
        columns = record
      } else {
        o := map[string]interface{}{}
        for i, n := range columns {
          o[n] = record[i]
        }
        //fmt.Printf("Proc: %s\n", o)
        for _, c := range procChan {
          c <- o
        }
      }
    }
	}

  log.Printf("Done Loading")
  for _, c := range procChan {
    close(c)
  }
  wg.Wait()

	return nil
}

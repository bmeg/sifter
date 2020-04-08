package transform

import (
  "os"
  "log"
  "fmt"
  "regexp"
  "strconv"
  "strings"
  "sync"

  "github.com/bmeg/sifter/evaluate"
  "encoding/json"

  "crypto/sha1"
  "encoding/csv"
  "github.com/bmeg/golib"
  "github.com/bmeg/sifter/pipeline"

)

type ObjectCreateStep struct {
  Class  string `json:"class"`
  Name   string `json:"name"`
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
  Mapping map[string]interface{} `json:"mapping"`
}

type FieldProcessStep struct {
  Column string                        `json:"col"`
  Steps   TransformPipe                `json:"steps"`
  Mapping map[string]string            `json:"mapping"`
  inChan  chan map[string]interface{}
}

type TableWriteStep struct {
  Output       string   `json:"output"`
  Columns      []string `json:"columns"`
  out          *os.File
  writer       *csv.Writer
}

type TableReplaceStep struct {
  Input        string   `json:"input"`
  Field        string   `json:"field"`
  table        map[string]string
}

type TableProjectStep struct {
  Input        string   `json:"input"`
  table        map[string]string
}


type DebugStep struct {
  Label        string                 `json:"label"`
}

type TransformStep struct {
  FieldMap      *FieldMapStep          `json:"fieldMap"`
  FieldType     *FieldTypeStep         `json:"fieldType"`
  //EdgeCreate    *EdgeCreateStep        `json:"edgeCreate"`
  //VertexCreate  *VertexCreateStep      `json:"vertexCreate"`
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
  TableReplace  *TableReplaceStep      `json:"tableReplace"`
  TableProject  *TableProjectStep      `json:"tableProject"`
  Fork          *ForkStep              `json:"fork"`
}

type TransformPipe []TransformStep



type ForkStep struct {
  Transform        []TransformPipe          `json:"transform"`
  pipes            []chan map[string]interface{}
}

func (fm FieldMapStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
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

func (fs FieldTypeStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
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

func (ts ObjectCreateStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
  task.Runtime.EmitObject(ts.Name, ts.Class, i)
  return i
}


func (tw *TableWriteStep) Start(task *pipeline.Task, wg *sync.WaitGroup) {
  path, _ := task.Path(tw.Output)
  tw.out, _ = os.Create(path)
  tw.writer = csv.NewWriter(tw.out)
  tw.writer.Comma = '\t'
  tw.writer.Write(tw.Columns)
}

func (tw *TableWriteStep)  Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
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

func (tr *TableReplaceStep) Start(task *pipeline.Task, wg *sync.WaitGroup) error {
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

func (tr *TableProjectStep) Start(task *pipeline.Task, wg *sync.WaitGroup) error {
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


func (fs *FilterStep) Start(task *pipeline.Task, wg *sync.WaitGroup) {
  fs.inChan = make(chan map[string]interface{}, 100)
  fs.Steps.Start(fs.inChan, task, wg)
}

func (fs FilterStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
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


func (fs *FieldProcessStep) Start(task *pipeline.Task, wg *sync.WaitGroup) {
  fs.inChan = make(chan map[string]interface{}, 100)
  fs.Steps.Start(fs.inChan, task, wg)
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
  for x,y := range i {
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

func (fs *ForkStep) Start(task *pipeline.Task, wg *sync.WaitGroup) error {
  fs.pipes = []chan map[string]interface{}{}
  for _, t := range fs.Transform {
    p := make(chan map[string]interface{}, 100)
    t.Start(p, task, wg)
    fs.pipes = append(fs.pipes, p)
  }
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
      /*
    } else if ts.VertexCreate != nil {
      for i := range in {
        out <- ts.VertexCreate.Run(i, task)
      }
    } else if ts.EdgeCreate != nil {
      for i := range in {
        out <- ts.EdgeCreate.Run(i, task)
      } */
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
        o := ts.Map.Run(i, task)
        out <- o
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
    } else if ts.TableReplace != nil {
      err := ts.TableReplace.Start(task, wg)
      if err != nil {
        log.Printf("TableReplace err: %s", err)
      }
      for i := range in {
        out <- ts.TableReplace.Run(i, task)
      }
    } else if ts.TableProject != nil {
      err := ts.TableProject.Start(task, wg)
      if err != nil {
        log.Printf("TableProject err: %s", err)
      }
      for i := range in {
        out <- ts.TableProject.Run(i, task)
      }
    } else if ts.Fork != nil {
      err := ts.Fork.Start(task, wg)
      if err != nil {
        log.Printf("Fork err: %s", err)
      }
      for i := range in {
        out <- ts.Fork.Run(i, task)
      }
      ts.Fork.Close()
    } else {
      log.Printf("Unknown field step")
    }
  }()
  return out
}

func (tp TransformPipe) Start( in chan map[string]interface{},
  task *pipeline.Task,
  wg *sync.WaitGroup) {

    log.Printf("Starting TransformPipe")
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

package emitter

import (
  "os"
  "io"
  "path"
  "log"
  "fmt"
  "sync"
  "compress/gzip"
  "encoding/json"
  "encoding/csv"
  "path/filepath"

  "github.com/bmeg/sifter/schema"
)

type DirEmitter struct {
  dir string
  mux sync.Mutex
  schemas *schema.Schemas
  vout map[string]io.WriteCloser
  eout map[string]io.WriteCloser
  oout map[string]io.WriteCloser
}

func NewDirEmitter(dir string, schemas *schema.Schemas) *DirEmitter {
  log.Printf("Emitting to %s", dir)
  return &DirEmitter{
    dir: dir,
    schemas:schemas,
    vout: map[string]io.WriteCloser{},
    eout: map[string]io.WriteCloser{},
    oout: map[string]io.WriteCloser{},
  }
}


func (s *DirEmitter) EmitObject(prefix string, objClass string, i map[string]interface{}) error {
  s.mux.Lock()
  defer s.mux.Unlock()
  name := fmt.Sprintf("%s.%s", prefix, objClass)
  f, ok := s.oout[name]
  if !ok {
    j, err := os.Create(path.Join(s.dir, name + ".json.gz" ))
    if err != nil {
      return err
    }
    f = gzip.NewWriter(j)
    s.oout[name] = f
  }
  v, err := s.schemas.Validate(objClass, i)
  if err != nil {
    log.Printf("Object Error: %s", err)
    return err
  }
  o, _ := json.Marshal(v)
  f.Write([]byte(o))
  f.Write([]byte("\n"))
  return nil
}

func (s *DirEmitter) Close() {
  log.Printf("Closing emitter")
  for _, v := range s.vout {
    v.Close()
  }
  for _, v := range s.eout {
    v.Close()
  }
  for _, v := range s.oout {
    v.Close()
  }
}


type dirTableEmitter struct {
  columns      []string
  out          io.WriteCloser
  handle       io.WriteCloser
  writer       *csv.Writer
}

func (s *dirTableEmitter) EmitRow(i map[string]interface{}) error {
  o := make([]string, len(s.columns))
  for j, k := range s.columns {
    if v, ok := i[k]; ok {
      if vStr, ok := v.(string); ok {
        o[j] = vStr
      }
    }
  }
  return s.writer.Write(o)
}


func (s *dirTableEmitter) Close() {
  log.Printf("Closing Table Writer")
  s.writer.Flush()
  s.out.Close()
  s.handle.Close()
}


func (s *DirEmitter) EmitTable( prefix string, columns []string ) TableEmitter {
  path := filepath.Join(s.dir, fmt.Sprintf("%s.table.gz", prefix))
  te := dirTableEmitter{}
  te.handle, _ = os.Create(path)
  te.out = gzip.NewWriter(te.handle)
  te.writer = csv.NewWriter(te.out)
  te.writer.Comma = '\t'
  te.columns = columns
  te.writer.Write(te.columns)
  return &te
}

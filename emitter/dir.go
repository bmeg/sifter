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

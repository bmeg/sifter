package manager

import (
  "os"
  "log"
  "io/ioutil"
  "path"
  "github.com/hashicorp/go-getter"
  "github.com/bmeg/grip/gripql"
  "github.com/bmeg/sifter/emitter"
)

type Manager struct {
  Workdir string
  Args    []string
  Output  emitter.Emitter
}


func Init(args []string) Manager {
  dir, err := ioutil.TempDir("./", "sifterwork_")
  if err != nil {
    log.Fatal(err)
  }
  s := emitter.StdoutEmitter{}
  return Manager{dir, args, s}
}

func (m Manager) Close() {
  os.RemoveAll(m.Workdir)
}

func (m Manager) Path(p string) string {
  return path.Join(m.Workdir, p)
}

func (m Manager) DownloadFile(url string) (string, error) {
  d := m.Path(path.Base(url))
  return d, getter.GetFile(d, url + "?archive=false")
}

func (m Manager) EmitVertex(v *gripql.Vertex) error {
  return m.Output.EmitVertex(v)
}

func (m Manager) EmitEdge(e *gripql.Edge) error {
  return m.Output.EmitEdge(e)
}

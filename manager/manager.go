package manager

import (
  "os"
  "log"
  "io/ioutil"
  "path"
  "github.com/hashicorp/go-getter"
)

type Manager struct {
  Workdir string
  Args    []string
}


func Init(args []string) Manager {
  dir, err := ioutil.TempDir("./", "sifterwork_")
  if err != nil {
    log.Fatal(err)
  }
  return Manager{dir, args}
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

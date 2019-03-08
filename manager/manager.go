package manager

import (
  "os"
  "log"
  "io/ioutil"
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

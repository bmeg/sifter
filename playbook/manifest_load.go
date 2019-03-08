package playbook

import (
  "log"
  "github.com/bmeg/grip/gripql"
  "github.com/bmeg/sifter/manager"
)


func (ml *ManifestLoadStep) Load(man manager.Manager) chan gripql.GraphElement {
  log.Printf("loading manifest %s", ml.Input)
  out := make(chan gripql.GraphElement, 10)
  close(out)
  return out
}

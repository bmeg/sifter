package playbook


import (
  "github.com/bmeg/grip/gripql"
)


func (ml *MatrixLoadStep) Load() chan gripql.GraphElement {
  out := make(chan gripql.GraphElement, 10)
  close(out)
  return out
}

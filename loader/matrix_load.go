package loader


import (
  "io"
  "github.com/bmeg/sifter/playbook"
  "github.com/bmeg/grip/gripql"
)

type Loader interface {
  Load(input io.Reader, config playbook.StepConfig) chan gripql.GraphElement
}

type MatrixLoader struct {}

func (ml MatrixLoader) Load(input io.Reader, config playbook.StepConfig) chan gripql.GraphElement {
  out := make(chan gripql.GraphElement, 10)
  close(out)
  return out
}

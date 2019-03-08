package emitter

import (
  "fmt"
)

type StdoutEmitter struct {

}

func (s StdoutEmitter) EmitVertex(v *gripql.Vertex) error {
  fmt.Printf("%s", v)
  return nil
}

func (s StdoutEmitter) EmitEdge(e *gripql.Edge) error {
  fmt.Printf("%s", e)
  return nil
}

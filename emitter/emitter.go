package emitter

import (
  "fmt"
  "github.com/bmeg/grip/gripql"
)

type Emitter interface {
  EmitVertex(v *gripql.Vertex) error
  EmitEdge(e *gripql.Edge) error
  Close()
}

type StdoutEmitter struct {

}

func (s StdoutEmitter) EmitVertex(v *gripql.Vertex) error {
  fmt.Printf("%s\n", v)
  return nil
}

func (s StdoutEmitter) EmitEdge(e *gripql.Edge) error {
  fmt.Printf("%s\n", e)
  return nil
}

func (s StdoutEmitter) Close() {}

package graph

import (
	"fmt"

	"github.com/bmeg/grip/gripql"
	"github.com/golang/protobuf/jsonpb"
)

type StdoutEmitter struct {
	jm jsonpb.Marshaler
}

func (s StdoutEmitter) EmitVertex(v *gripql.Vertex) error {
	o, _ := s.jm.MarshalToString(v)
	fmt.Printf("%s\n", o)
	return nil
}

func (s StdoutEmitter) EmitEdge(e *gripql.Edge) error {
	o, _ := s.jm.MarshalToString(e)
	fmt.Printf("%s\n", o)
	return nil
}

func (s StdoutEmitter) Close() {}

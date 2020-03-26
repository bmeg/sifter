package emitter

import (
	"fmt"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"
)

type StdoutEmitter struct {
	jm jsonpb.Marshaler
	storeObjects bool
	schemas schema.Schemas
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

func (s StdoutEmitter) EmitObject(objClass string, i map[string]interface{}) error {
	if s.storeObjects {
		o, _ := json.Marshal(i)
		fmt.Printf("%s : %s\n", objClass, o)
	} else {
		return GenerateGraph(s.schemas, objClass, i, s)
	}
	return nil
}


func (s StdoutEmitter) Close() {}

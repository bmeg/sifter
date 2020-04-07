package emitter

import (
	"fmt"
	"log"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"
)

type StdoutEmitter struct {
	jm jsonpb.Marshaler
	schemas *schema.Schemas
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
	v, err := s.schemas.Validate(objClass, i)
	if err != nil {
		log.Printf("Object Error: %s", err)
		return err
	}
	o, _ := json.Marshal(v)
	fmt.Printf("%s : %s\n", objClass, o)
	return nil
}


func (s StdoutEmitter) Close() {}

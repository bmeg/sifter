package loader

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bmeg/sifter/schema"

	"github.com/bmeg/grip/gripql"
	"github.com/golang/protobuf/jsonpb"
)

type StdoutLoader struct { }

func (s StdoutLoader) Close() {}

func (s StdoutLoader) NewDataEmitter(schemas *schema.Schemas) (DataEmitter, error) {
	return StdoutEmitter{schemas:schemas}, nil
}

func (s StdoutLoader) NewGraphEmitter() (GraphEmitter, error) {
	return StdoutEmitter{}, nil
}


type StdoutEmitter struct {
	jm      jsonpb.Marshaler
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


func (s StdoutEmitter) Emit(name string, v map[string]interface{}) error {
	o, _ := json.Marshal(v)
	fmt.Printf("%s\t%s\n", name, o)
	return nil
}

func (s StdoutEmitter) EmitObject(prefix string, objClass string, i map[string]interface{}) error {
	v, err := s.schemas.Validate(objClass, i)
	if err != nil {
		log.Printf("Object Error: %s", err)
		return err
	}
	o, _ := json.Marshal(v)
	fmt.Printf("%s.%s : %s\n", prefix, objClass, o)
	return nil
}

type stdTableEmitter struct {
	columns []string
}

func (s *stdTableEmitter) EmitRow(i map[string]interface{}) error {
	o := make([]string, len(s.columns))
	for j, k := range s.columns {
		if v, ok := i[k]; ok {
			if vStr, ok := v.(string); ok {
				o[j] = vStr
			}
		}
	}
	fmt.Printf("%#v\n", o)
	return nil
}

func (s *stdTableEmitter) Close() {}

func (s StdoutEmitter) EmitTable(prefix string, columns []string, sep rune) TableEmitter {
	te := stdTableEmitter{columns}
	fmt.Printf("%s\n", columns)
	return &te
}

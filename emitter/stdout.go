package emitter

import (
	"fmt"
	"log"
	"encoding/json"
	"github.com/bmeg/sifter/schema"
)

type StdoutEmitter struct {
	schemas *schema.Schemas
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


func (s StdoutEmitter) Close() {}


type stdTableEmitter struct {
  columns      []string
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

func (s StdoutEmitter) EmitTable( prefix string, columns []string ) TableEmitter {
 	te := stdTableEmitter{columns}
	fmt.Printf("%s\n", columns)
  return &te
}

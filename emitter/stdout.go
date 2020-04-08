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

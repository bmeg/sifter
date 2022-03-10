package loader

import (
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
)

type StdoutLoader struct{}

func (s StdoutLoader) Close() {}

func (s StdoutLoader) NewDataEmitter() (DataEmitter, error) {
	return StdoutEmitter{}, nil
}

type StdoutEmitter struct {
	jm jsonpb.Marshaler
}

func (s StdoutEmitter) Close() {}

func (s StdoutEmitter) Emit(name string, v map[string]interface{}) error {
	o, _ := json.Marshal(v)
	fmt.Printf("%s\t%s\n", name, o)
	return nil
}

package emitter

import (
	"fmt"
	"log"
	"net/url"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"
)

type Emitter interface {
	EmitVertex(v *gripql.Vertex) error
	EmitEdge(e *gripql.Edge) error
	EmitObject(objClass string, e map[string]interface{}) error
	Close()
}

func GraphExists(server string, graph string) (bool, error) {
	u, _ := url.Parse(server)

	if u.Scheme == "grip" {
		log.Printf("Checking %s for %s", u.Host, graph)
		return GripGraphExists(u.Host, graph)
	}
	if u.Scheme == "mongodb" {
		return MongoGraphExists(server, graph)
	}
	if u.Scheme == "stdout" {
		return false, nil
	}
	if u.Scheme == "dir" {
		return false, nil
	}
	return false, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

func NewEmitter(server string, graph string, storeObjects bool, sc *schema.Schemas) (Emitter, error) {
	u, _ := url.Parse(server)
	if u.Scheme == "grip" {
		return NewGripEmitter(u.Host, graph)
	}
	if u.Scheme == "mongodb" {
		return NewMongoEmitter(server, graph)
	}
	if u.Scheme == "stdout" {
		return StdoutEmitter{storeObjects:storeObjects}, nil
	}
	if u.Scheme == "dir" {
		return NewDirEmitter( u.Host + u.Path, storeObjects, sc ), nil
	}
	return nil, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

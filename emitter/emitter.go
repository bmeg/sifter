package emitter

import (
	"fmt"
	"log"
	"net/url"

	"github.com/bmeg/grip/gripql"
)

type Emitter interface {
	EmitVertex(v *gripql.Vertex) error
	EmitEdge(e *gripql.Edge) error
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
	return false, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

func NewEmitter(server string, graph string) (Emitter, error) {
	u, _ := url.Parse(server)
	if u.Scheme == "grip" {
		return NewGripEmitter(u.Host, graph)
	}
	if u.Scheme == "mongodb" {
		return NewMongoEmitter(server, graph)
	}
	return nil, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

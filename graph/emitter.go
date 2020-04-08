package graph

import (
	"fmt"
	"log"
	"net/url"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"
)

type GraphEmitter interface {
	EmitVertex(v *gripql.Vertex) error
	EmitEdge(e *gripql.Edge) error
	Close()
}

func GraphExists(server string, graph string, args string) (bool, error) {
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

func NewGraphEmitter(driver string, sc *schema.Schemas) (GraphEmitter, error) {
	u, _ := url.Parse(driver)
	if u.Scheme == "grip" {
		return NewGripEmitter(u.Host, u.Path)
	}
	if u.Scheme == "mongodb" {
		return NewMongoEmitter(u.Host, u.Path)
	}
	if u.Scheme == "stdout" {
		return StdoutEmitter{}, nil
	}
	if u.Scheme == "dir" {
		return NewDirEmitter( u.Host + u.Path, sc ), nil
	}
	return nil, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

package loader

import (
	"fmt"
	"log"
	"net/url"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"
)

type TableEmitter interface {
	EmitRow(values map[string]interface{}) error
	Close()
}

type DataEmitter interface {
	Emit(name string, e map[string]interface{}) error
	EmitObject(prefix string, objClass string, e map[string]interface{}) error
	EmitTable(prefix string, columns []string, sep rune) TableEmitter
}

type GraphEmitter interface {
	EmitVertex(v *gripql.Vertex) error
	EmitEdge(e *gripql.Edge) error
}

type Loader interface {
	NewDataEmitter(*schema.Schemas) (DataEmitter, error)
	NewGraphEmitter() (GraphEmitter, error)
	Close()
}


func NewLoader(driver string) (Loader, error) {
	u, _ := url.Parse(driver)
	if u.Scheme == "stdout" {
		return StdoutLoader{}, nil
	}
	if u.Scheme == "dir" {
		return NewDirLoader(u.Host+u.Path), nil
	}
	if u.Scheme == "grip" {
		return NewGripLoader(u.Host, u.Path)
	}
	if u.Scheme == "mongodb" {
		return NewMongoLoader(u.Host, u.Path)
	}
	return nil, fmt.Errorf("Unknown driver: %s", u.Scheme)
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

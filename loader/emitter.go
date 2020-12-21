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
	Close()
}

type GraphEmitter interface {
	EmitVertex(v *gripql.Vertex) error
	EmitEdge(e *gripql.Edge) error
	Close()
}

func NewDataEmitter(driver string, sc *schema.Schemas) (DataEmitter, error) {
	u, _ := url.Parse(driver)
	if u.Scheme == "stdout" {
		return StdoutEmitter{schemas: sc}, nil
	}
	if u.Scheme == "dir" {
		return NewDirEmitter(u.Host+u.Path, sc), nil
	}
	return nil, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

func NewGraphEmitter(driver string) (GraphEmitter, error) {
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
		return NewDirEmitter(u.Host+u.Path, nil), nil
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

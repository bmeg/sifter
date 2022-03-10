package loader

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"
)

type DataEmitter interface {
	Emit(name string, e map[string]interface{}) error
	EmitObject(prefix string, objClass string, e map[string]interface{}) error
	Close()
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
		return NewDirLoader(u.Host + u.Path), nil
	}
	if u.Scheme == "grip" {
		if strings.HasPrefix(u.Path, "/") {
			u.Path = u.Path[1:len(u.Path)]
		}
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

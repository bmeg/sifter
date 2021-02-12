package graphbuild

import (
	"log"
	"path/filepath"

	"github.com/akrylysov/pogreb"
)

type Check struct {
	vertDB *pogreb.DB
	edgeDB *pogreb.DB
}

func NewGraphCheck(dir string) (*Check, error) {
	vertFile := filepath.Join(dir, "vert.db")
	edgeFile := filepath.Join(dir, "edge.db")
	vertDB, err := pogreb.Open(vertFile, nil)
	if err != nil {
		return nil, err
	}
	edgeDB, err := pogreb.Open(edgeFile, nil)
	if err != nil {
		return nil, err
	}
	return &Check{vertDB: vertDB, edgeDB: edgeDB}, nil
}

func (gc *Check) AddVertex(g string) {
	gc.vertDB.Put([]byte(g), []byte{})
}

func (gc *Check) AddEdge(from, to string) {
	gc.edgeDB.Put([]byte(from), []byte(to))
	gc.edgeDB.Put([]byte(to), []byte(from))
}

func (gc *Check) GetEdgeVertices() chan string {
	out := make(chan string, 100)
	go func() {
		defer close(out)
		it := gc.edgeDB.Items()
		for {
			key, _, err := it.Next()
			if err == pogreb.ErrIterationDone {
				break
			}
			if err != nil {
				log.Printf("Edge DB Error: %s", err)
				return
			}
			out <- string(key)
		}
	}()
	return out
}

func (gc *Check) HasVertex(s string) bool {
	val, err := gc.vertDB.Get([]byte(s))
	if val == nil || err != nil {
		return false
	}
	return true
}

func (gc *Check) GetEdgeSource(s string) string {
	val, err := gc.edgeDB.Get([]byte(s))
	if val == nil || err != nil {
		return ""
	}
	return string(val)
}

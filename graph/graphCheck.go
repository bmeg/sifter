package graph

import (
	"log"
	"path/filepath"

	"github.com/akrylysov/pogreb"
)

type GraphCheck struct {
	vertDB *pogreb.DB
	edgeDB *pogreb.DB
}

func NewGraphCheck(dir string) (*GraphCheck, error) {
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
	return &GraphCheck{vertDB: vertDB, edgeDB: edgeDB}, nil
}

func (gc *GraphCheck) AddVertex(g string) {
	gc.vertDB.Put([]byte(g), []byte{})
}

func (gc *GraphCheck) AddEdge(from, to string) {
	gc.edgeDB.Put([]byte(from), []byte(to))
	gc.edgeDB.Put([]byte(to), []byte(from))
}

func (gc *GraphCheck) GetEdgeVertices() chan string {
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

func (gc *GraphCheck) HasVertex(s string) bool {
	val, err := gc.vertDB.Get([]byte(s))
	if val == nil || err != nil {
		return false
	}
	return true
}

func (gc *GraphCheck) GetEdgeSource(s string) string {
	val, err := gc.edgeDB.Get([]byte(s))
	if val == nil || err != nil {
		return ""
	}
	return string(val)
}

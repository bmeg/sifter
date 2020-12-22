package loader

import (
	"log"
	"fmt"
	"sync"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/grip/util/rpc"
	"github.com/bmeg/sifter/schema"
)

type GripLoader struct {
	client   gripql.Client
	graph    string
	elemChan chan *gripql.GraphElement
	done     *sync.WaitGroup
}

func GripGraphExists(host string, graph string) (bool, error) {
	conn, err := gripql.Connect(rpc.ConfigWithDefaults(host), true)
	if err != nil {
		return false, err
	}

	resp, err := conn.ListGraphs()
	if err != nil {
		return false, err
	}

	found := false
	for _, g := range resp.Graphs {
		if graph == g {
			found = true
		}
	}
	conn.Close()
	return found, nil
}

// NewGripEmitter
func NewGripLoader(host string, graph string) (*GripLoader, error) {

	conn, err := gripql.Connect(rpc.ConfigWithDefaults(host), true)
	if err != nil {
		return nil, err
	}

	resp, err := conn.ListGraphs()
	if err != nil {
		return nil, err
	}

	found := false
	for _, g := range resp.Graphs {
		if graph == g {
			found = true
		}
	}
	if !found {
		log.Printf("creating graph: %s", graph)
		err := conn.AddGraph(graph)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("loading graph: %s", graph)
	elemChan := make(chan *gripql.GraphElement, 1000)
	done := sync.WaitGroup{}
	done.Add(1)
	go loadFunc(conn, elemChan, &done)

	return &GripLoader{conn, graph, elemChan, &done}, nil

}


func (s *GripLoader) NewDataEmitter(sc *schema.Schemas) (DataEmitter, error) {
	return nil, fmt.Errorf("GRIP data loader not implemented")
}

func (s *GripLoader) NewGraphEmitter() (GraphEmitter, error) {
	return s, nil
}

func loadFunc(conn gripql.Client, elemChan chan *gripql.GraphElement, done *sync.WaitGroup) {
	if err := conn.BulkAdd(elemChan); err != nil {
		log.Printf("bulk add error: %v", err)
	}
	log.Printf("Bulk Write done")
	done.Done()
}

func (s *GripLoader) EmitVertex(v *gripql.Vertex) error {
	s.elemChan <- &gripql.GraphElement{Graph: s.graph, Vertex: v}
	return nil
}

func (s *GripLoader) EmitEdge(e *gripql.Edge) error {
	s.elemChan <- &gripql.GraphElement{Graph: s.graph, Edge: e}
	return nil
}

func (s *GripLoader) Close() {
	log.Printf("Closing GRIP connection")
	close(s.elemChan)
	s.done.Wait()
	log.Printf("Closed GRIP connection")
}

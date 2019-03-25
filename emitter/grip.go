package emitter

import (
	"log"
	"sync"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/grip/util/rpc"
)

type GripEmitter struct {
	client   gripql.Client
	graph    string
	elemChan chan *gripql.GraphElement
	done     sync.WaitGroup
}

// NewGripEmitter
func NewGripEmitter(host string, graph string) (GripEmitter, error) {

	conn, err := gripql.Connect(rpc.ConfigWithDefaults(host), true)
	if err != nil {
		return GripEmitter{}, err
	}

	resp, err := conn.ListGraphs()
	if err != nil {
		return GripEmitter{}, err
	}

	found := false
	for _, g := range resp.Graphs {
		if graph == g {
			found = true
		}
	}
	if !found {
		log.Printf("creating graph")
		err := conn.AddGraph(graph)
		if err != nil {
			return GripEmitter{}, err
		}
	}

	elemChan := make(chan *gripql.GraphElement)
	done := sync.WaitGroup{}
	done.Add(1)
	go loadFunc(conn, elemChan, done)

	return GripEmitter{conn, graph, elemChan, done}, nil

}

func loadFunc(conn gripql.Client, elemChan chan *gripql.GraphElement, done sync.WaitGroup) {
	if err := conn.BulkAdd(elemChan); err != nil {
		log.Printf("bulk add error: %v", err)
	}
	log.Printf("Bulk Write done")
	done.Done()
}

func (s GripEmitter) EmitVertex(v *gripql.Vertex) error {
	s.elemChan <- &gripql.GraphElement{Graph: s.graph, Vertex: v}
	return nil
}

func (s GripEmitter) EmitEdge(e *gripql.Edge) error {
	s.elemChan <- &gripql.GraphElement{Graph: s.graph, Edge: e}
	return nil
}

func (s GripEmitter) Close() {
	close(s.elemChan)
	s.done.Wait()
}

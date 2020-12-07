package graph

import (
	"compress/gzip"
	"io"
	"log"
	"os"
	"path"
	"sync"

	"github.com/bmeg/grip/gripql"
	"github.com/golang/protobuf/jsonpb"
)

type DirEmitter struct {
	jm   jsonpb.Marshaler
	dir  string
	mux  sync.Mutex
	vout map[string]io.WriteCloser
	eout map[string]io.WriteCloser
	oout map[string]io.WriteCloser
}

func NewDirEmitter(dir string) *DirEmitter {
	log.Printf("Emitting to %s", dir)
	return &DirEmitter{
		jm:   jsonpb.Marshaler{},
		dir:  dir,
		vout: map[string]io.WriteCloser{},
		eout: map[string]io.WriteCloser{},
		oout: map[string]io.WriteCloser{},
	}
}

func (s *DirEmitter) EmitVertex(v *gripql.Vertex) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	f, ok := s.vout[v.Label]
	if !ok {
		j, err := os.Create(path.Join(s.dir, v.Label+".Vertex.json.gz"))
		if err != nil {
			log.Printf("Error: %s", err)
			return err
		}
		f = gzip.NewWriter(j)
		s.vout[v.Label] = f
	}
	o, _ := s.jm.MarshalToString(v)
	f.Write([]byte(o))
	f.Write([]byte("\n"))
	return nil
}

func (s *DirEmitter) EmitEdge(e *gripql.Edge) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	f, ok := s.eout[e.Label]
	if !ok {
		j, err := os.Create(path.Join(s.dir, e.Label+".Edge.json.gz"))
		if err != nil {
			return err
		}
		f = gzip.NewWriter(j)
		s.eout[e.Label] = f
	}
	o, err := s.jm.MarshalToString(e)
	if err != nil {
		log.Printf("Error: %s", err)
		return err
	}
	f.Write([]byte(o))
	f.Write([]byte("\n"))
	return nil
}

func (s *DirEmitter) Close() {
	log.Printf("Closing emitter")
	for _, v := range s.vout {
		v.Close()
	}
	for _, v := range s.eout {
		v.Close()
	}
	for _, v := range s.oout {
		v.Close()
	}
}

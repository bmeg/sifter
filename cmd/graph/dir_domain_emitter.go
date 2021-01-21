package graph


import (
	"compress/gzip"
  "strings"
	"io"
	"log"
	"os"
	"path"
	"sync"

	"github.com/bmeg/grip/gripql"
	"github.com/golang/protobuf/jsonpb"
)

type GraphDomainEmitter struct {
  jm   jsonpb.Marshaler
  dir  string
  mux  sync.Mutex
  vertDomains []string
  edgeEndDomains [][]string
  vout map[string]io.WriteCloser
  eout map[string]io.WriteCloser
}

func NewGraphDomainEmitter(baseDir string, vertDomains []string, edgeEndDomains [][]string) *GraphDomainEmitter {
  return &GraphDomainEmitter{
    dir:baseDir,
    jm:jsonpb.Marshaler{},
    vertDomains:vertDomains,
    edgeEndDomains:edgeEndDomains,
    vout: map[string]io.WriteCloser{},
    eout: map[string]io.WriteCloser{},
  }
}

func (s *GraphDomainEmitter) EmitVertex(v *gripql.Vertex) error {
	s.mux.Lock()
	defer s.mux.Unlock()

  vDomain := s.getVertDomain(v)

  prefix := vDomain + "." + v.Label

	f, ok := s.vout[prefix]
	if !ok {
		j, err := os.Create(path.Join(s.dir, prefix + ".Vertex.json.gz"))
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

func (s *GraphDomainEmitter) EmitEdge(e *gripql.Edge) error {
	s.mux.Lock()
	defer s.mux.Unlock()

  eDomain := s.getEdgeEndDomain( e )
  prefix := eDomain + "." + e.Label

	f, ok := s.eout[prefix]
	if !ok {
		j, err := os.Create(path.Join(s.dir, prefix + ".Edge.json.gz"))
		if err != nil {
			return err
		}
		f = gzip.NewWriter(j)
		s.eout[prefix] = f
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


func (s *GraphDomainEmitter) Close() {
	log.Printf("Closing emitter")
	for _, v := range s.vout {
		v.Close()
	}
	for _, v := range s.eout {
		v.Close()
	}
}


func (s *GraphDomainEmitter) getVertDomain(v *gripql.Vertex) string {
  for _, i := range s.vertDomains {
    if strings.HasPrefix(v.Gid, i) {
      return i
    }
  }
  return ""
}

func (s *GraphDomainEmitter) getEdgeEndDomain(e *gripql.Edge) string {
  for _, i := range s.edgeEndDomains {
    if strings.HasPrefix(e.From, i[0]) && strings.HasPrefix(e.To, i[1]) {
      return i[0] + "_" + i[1]
    }
  }
  return ""
}

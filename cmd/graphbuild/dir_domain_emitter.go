package graphbuild

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/bmeg/grip/gripql"
	"github.com/golang/protobuf/jsonpb"
)

type DomainEmitter struct {
	jm             jsonpb.Marshaler
	dir            string
	filePrefix     string
	mux            sync.Mutex
	vertDomains    []string
	edgeEndDomains [][]string
	vout           map[string]io.WriteCloser
	eout           map[string]io.WriteCloser
}

func NewDomainEmitter(baseDir string, prefix string, vertDomains []string, edgeEndDomains [][]string) *DomainEmitter {
	return &DomainEmitter{
		dir:            baseDir,
		filePrefix:     prefix,
		jm:             jsonpb.Marshaler{},
		vertDomains:    vertDomains,
		edgeEndDomains: edgeEndDomains,
		vout:           map[string]io.WriteCloser{},
		eout:           map[string]io.WriteCloser{},
	}
}

func (s *DomainEmitter) EmitVertex(v *gripql.Vertex) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	vDomain := s.getVertDomain(v)
	if vDomain == "" {
		return fmt.Errorf("Domain for %s not found", v)
	}

	var prefix string
	if s.filePrefix == "" {
		prefix = vDomain + "." + v.Label
	} else {
		prefix = s.filePrefix + "." + v.Label
	}

	f, ok := s.vout[prefix]
	if !ok {
		j, err := os.Create(path.Join(s.dir, prefix+".Vertex.json.gz"))
		if err != nil {
			log.Printf("Error: %s", err)
			return err
		}
		f = gzip.NewWriter(j)
		s.vout[prefix] = f
	}
	o, _ := s.jm.MarshalToString(v)
	f.Write([]byte(o))
	f.Write([]byte("\n"))
	return nil
}

func (s *DomainEmitter) EmitEdge(e *gripql.Edge) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	eDomain := s.getEdgeEndDomain(e)
	if eDomain == "" {
		return fmt.Errorf("Edge Prefix not found for %s", e.Label)
	}

	var prefix string
	if s.filePrefix == "" {
		prefix = eDomain + "." + e.Label
	} else {
		prefix = s.filePrefix + "." + eDomain + "." + e.Label
	}
	f, ok := s.eout[prefix]
	if !ok {
		j, err := os.Create(path.Join(s.dir, prefix+".Edge.json.gz"))
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

func (s *DomainEmitter) Close() {
	log.Printf("Closing emitter")
	for _, v := range s.vout {
		v.Close()
	}
	for _, v := range s.eout {
		v.Close()
	}
}

func (s *DomainEmitter) getVertDomain(v *gripql.Vertex) string {
	for _, i := range s.vertDomains {
		if strings.HasPrefix(v.Gid, i) {
			return i
		}
	}
	return ""
}

func (s *DomainEmitter) getEdgeEndDomain(e *gripql.Edge) string {
	for _, i := range s.edgeEndDomains {
		if strings.HasPrefix(e.From, i[0]) && strings.HasPrefix(e.To, i[1]) {
			return i[0] + "_" + i[1]
		}
	}
	log.Printf("Can't find prefix match for edge (%s)-%s>(%s)", e.From, e.Label, e.To)
	return ""
}

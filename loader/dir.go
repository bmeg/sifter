package loader

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"
	"github.com/golang/protobuf/jsonpb"
)

type DirLoader struct {
	jm   jsonpb.Marshaler
	dir  string
	mux  sync.Mutex
	vout map[string]io.WriteCloser
	eout map[string]io.WriteCloser
	oout map[string]io.WriteCloser
	dout map[string]io.WriteCloser
}

type DirDataLoader struct {
	schemas *schema.Schemas
	dl      *DirLoader
}

func (s *DirLoader) NewGraphEmitter() (GraphEmitter, error) {
	return s, nil
}

func (s *DirLoader) NewDataEmitter(sc *schema.Schemas) (DataEmitter, error) {
	return &DirDataLoader{sc, s}, nil
}

func NewDirLoader(dir string) *DirLoader {
	log.Printf("Emitting to %s", dir)
	return &DirLoader{
		jm:   jsonpb.Marshaler{},
		dir:  dir,
		vout: map[string]io.WriteCloser{},
		eout: map[string]io.WriteCloser{},
		oout: map[string]io.WriteCloser{},
		dout: map[string]io.WriteCloser{},
	}
}

func (s *DirLoader) EmitVertex(v *gripql.Vertex) error {
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

func (s *DirLoader) EmitEdge(e *gripql.Edge) error {
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

func (s *DirLoader) Close() {
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
	for _, v := range s.dout {
		v.Close()
	}
}

func (s *DirDataLoader) Emit(name string, v map[string]interface{}) error {
	s.dl.mux.Lock()
	defer s.dl.mux.Unlock()
	f, ok := s.dl.dout[name]
	if !ok {
		j, err := os.Create(path.Join(s.dl.dir, name+".json.gz"))
		if err != nil {
			return err
		}
		f = gzip.NewWriter(j)
		s.dl.dout[name] = f
	}
	o, _ := json.Marshal(v)
	f.Write([]byte(o))
	f.Write([]byte("\n"))
	return nil
}

func (s *DirDataLoader) EmitObject(prefix string, objClass string, i map[string]interface{}) error {
	s.dl.mux.Lock()
	defer s.dl.mux.Unlock()
	name := fmt.Sprintf("%s.%s", prefix, objClass)
	f, ok := s.dl.oout[name]
	if !ok {
		j, err := os.Create(path.Join(s.dl.dir, name+".json.gz"))
		if err != nil {
			return err
		}
		f = gzip.NewWriter(j)
		s.dl.oout[name] = f
	}
	v, err := s.schemas.Validate(objClass, i)
	if err != nil {
		log.Printf("Object Error: %s", err)
		return err
	}
	o, _ := json.Marshal(v)
	f.Write([]byte(o))
	f.Write([]byte("\n"))
	return nil
}

type dirTableEmitter struct {
	columns []string
	out     io.WriteCloser
	handle  io.WriteCloser
	writer  *csv.Writer
}

func (s *dirTableEmitter) EmitRow(i map[string]interface{}) error {
	o := make([]string, len(s.columns))
	for j, k := range s.columns {
		if v, ok := i[k]; ok {
			if vStr, ok := v.(string); ok {
				o[j] = vStr
			}
		}
	}
	return s.writer.Write(o)
}

func (s *dirTableEmitter) Close() {
	log.Printf("Closing Table Writer")
	s.writer.Flush()
	s.out.Close()
	s.handle.Close()
}

func (s *DirDataLoader) EmitTable(name string, columns []string, sep rune) TableEmitter {
	path := filepath.Join(s.dl.dir, name)
	te := dirTableEmitter{}
	te.handle, _ = os.Create(path)
	if strings.HasSuffix(name, ".gz") {
		te.out = gzip.NewWriter(te.handle)
	} else {
		te.out = te.handle
	}
	te.writer = csv.NewWriter(te.out)
	te.writer.Comma = sep
	te.columns = columns
	te.writer.Write(te.columns)
	return &te
}

package loader

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
)

type DirLoader struct {
	jm   protojson.MarshalOptions
	dir  string
	mux  sync.Mutex
	vout map[string]io.WriteCloser
	eout map[string]io.WriteCloser
	oout map[string]io.WriteCloser
	dout map[string]io.WriteCloser
}

type DirDataLoader struct {
	dl *DirLoader
}

func (s *DirLoader) NewDataEmitter() (DataEmitter, error) {
	return &DirDataLoader{s}, nil
}

func NewDirLoader(dir string) *DirLoader {
	dir, _ = filepath.Abs(dir)
	//log.Printf("Emitting to %s", dir)
	return &DirLoader{
		jm:   protojson.MarshalOptions{},
		dir:  dir,
		vout: map[string]io.WriteCloser{},
		eout: map[string]io.WriteCloser{},
		oout: map[string]io.WriteCloser{},
		dout: map[string]io.WriteCloser{},
	}
}

func (s *DirLoader) Close() {
	//log.Printf("Closing emitter")
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
	s.vout = map[string]io.WriteCloser{}
	s.eout = map[string]io.WriteCloser{}
	s.oout = map[string]io.WriteCloser{}
	s.dout = map[string]io.WriteCloser{}
}

func (s *DirDataLoader) Emit(name string, v map[string]interface{}, useName bool) error {
	s.dl.mux.Lock()
	defer s.dl.mux.Unlock()
	f, ok := s.dl.dout[name]
	if !ok {
		// log.Printf("output path %s", outputPath)

		j, err := os.Create(path.Join(s.dl.dir, name+".json.gz"))
		if err != nil {
			return err
		}
		f = gzip.NewWriter(j)
		s.dl.dout[name] = f
	}
	if v != nil {
		o, _ := json.Marshal(v)
		f.Write([]byte(o))
		f.Write([]byte("\n"))
	}
	return nil
}

func (s *DirDataLoader) Close() {
	s.dl.Close()
}

package loader

import (
	"os"

	"github.com/biogo/hts/bgzf"
	"github.com/bmeg/grip/gripql"
	"google.golang.org/protobuf/encoding/protojson"
)

type BgzipGraphWriter struct {
	writer *bgzf.Writer
}

func NewBGZipGraphEmitter(path string) (*BgzipGraphWriter, error) {
	j, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	f := bgzf.NewWriter(j, 0)
	return &BgzipGraphWriter{f}, nil
}

func (bw *BgzipGraphWriter) Close() {
	bw.writer.Close()
}

func (bw *BgzipGraphWriter) EmitVertex(v *gripql.Vertex) error {
	o, err := protojson.Marshal(v)
	if err != nil {
		return err
	}
	bw.writer.Write(o)
	bw.writer.Write([]byte("\n"))
	return nil
}

func (bw *BgzipGraphWriter) EmitEdge(e *gripql.Edge) error {
	o, err := protojson.Marshal(e)
	if err != nil {
		return err
	}
	bw.writer.Write(o)
	bw.writer.Write([]byte("\n"))
	return nil
}

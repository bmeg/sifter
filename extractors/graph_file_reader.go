package extractors

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bmeg/golib"
	"github.com/bmeg/grip/gripql"
	"google.golang.org/protobuf/encoding/protojson"
)

func LoadVertexFile(path string) (chan *gripql.Vertex, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	log.Printf("Loading: %s", path)

	var reader chan []byte
	var err error
	if strings.HasSuffix(path, ".gz") {
		reader, err = golib.ReadGzipLines(path)
	} else {
		reader, err = golib.ReadFileLines(path)
	}
	if err != nil {
		return nil, err
	}

	out := make(chan *gripql.Vertex, 10)

	go func() {
		um := protojson.UnmarshalOptions{DiscardUnknown: true}
		defer close(out)
		for line := range reader {
			if len(line) > 0 {
				o := gripql.Vertex{}
				err := um.Unmarshal(line, &o)
				if err == nil {
					out <- &o
				}
			}
		}
	}()
	return out, nil
}

func LoadEdgeFile(path string) (chan *gripql.Edge, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	log.Printf("Loading: %s", path)

	var reader chan []byte
	var err error
	if strings.HasSuffix(path, ".gz") {
		reader, err = golib.ReadGzipLines(path)
	} else {
		reader, err = golib.ReadFileLines(path)
	}
	if err != nil {
		return nil, err
	}

	out := make(chan *gripql.Edge, 10)
	go func() {
		defer close(out)
		um := protojson.UnmarshalOptions{DiscardUnknown: true}
		for line := range reader {
			if len(line) > 0 {
				o := gripql.Edge{}
				err := um.Unmarshal(line, &o)
				if err == nil {
					out <- &o
				} else {
					fmt.Printf("Parse Error: %s\n", err)
				}
			}
		}
	}()
	return out, nil
}

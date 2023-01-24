package extractors

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"

	xj "github.com/basgys/goxml2json"
)

type XMLLoadStep struct {
	Input string `json:"input"`
	Level int    `json:"level"`
}

func (ml *XMLLoadStep) Start(task task.RuntimeTask) (chan map[string]any, error) {
	//log.Printf("Starting XML Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", input)
	}
	log.Printf("Loading: %s", input)

	fhd, err := os.Open(input)
	if err != nil {
		return nil, err
	}

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return nil, err
		}
	} else {
		hd = fhd
	}

	procChan := make(chan map[string]any, 100)
	if ml.Level == 0 {
		go func() {
			jStr, err := xj.Convert(hd)
			if err == nil {
				data := map[string]any{}
				if err = json.Unmarshal(jStr.Bytes(), &data); err == nil {
					procChan <- data
				}
			}
			close(procChan)
		}()
	} else {
		go func() {
			d := xml.NewDecoder(hd)
			stack := []string{}
			buffer := []xml.Token{}
			for {
				tok, err := d.Token()
				if tok == nil || err == io.EOF {
					// EOF means we're done.
					break
				} else if err != nil {
					log.Printf("Error decoding token: %s", err)
				}
				if len(stack) >= ml.Level {
					buffer = append(buffer, xml.CopyToken(tok))
				}
				switch ty := tok.(type) {
				case xml.StartElement:
					stack = append(stack, ty.Name.Local)
				case xml.EndElement:
					stack = stack[:len(stack)-1]
					if len(stack) == ml.Level {
						b := &bytes.Buffer{}
						e := xml.NewEncoder(b)
						for _, i := range buffer {
							e.EncodeToken(i)
						}
						e.Flush()
						jStr, err := xj.Convert(b)
						if err == nil {
							data := map[string]any{}
							if err = json.Unmarshal(jStr.Bytes(), &data); err == nil {
								procChan <- data
							}
						}
						buffer = []xml.Token{}
					}
				default:
				}
			}
			close(procChan)
		}()
	}
	return procChan, nil
}

func (ml *XMLLoadStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(ml.Input) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

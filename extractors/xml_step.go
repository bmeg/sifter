package extractors

import (
	"bufio"
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
}

func xmlStream(file io.Reader, out chan map[string]interface{}) {
	buffer := bufio.NewReaderSize(file, 1024*1024*256) // 33554432
	decoder := xml.NewDecoder(buffer)

	nameStack := []string{}
	mapStack := []map[string]interface{}{}
	curString := []byte{}
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			nameStack = append(nameStack, se.Name.Local)
			mapStack = append(mapStack, map[string]interface{}{})
			curString = []byte{}
			attributes := map[string][]string{}
			for _, a := range se.Attr {
				if x, ok := attributes[a.Name.Local]; ok {
					attributes[a.Name.Local] = append(x, a.Value)
				} else {
					attributes[a.Name.Local] = []string{a.Value}
				}
			}
			mattributes := map[string]any{}
			for k, v := range attributes {
				mattributes[k] = v
			}
			if len(mattributes) > 0 {
				mapStack[len(mapStack)-1]["_attr"] = mattributes
			}
		case xml.EndElement:
			cMap := mapStack[len(mapStack)-1]
			nameStack = nameStack[0 : len(nameStack)-1]
			mapStack = mapStack[0 : len(mapStack)-1]
			if len(mapStack) > 0 {
				if len(cMap) > 0 {
					//the child structure contained substructures
					if a, ok := mapStack[len(mapStack)-1][se.Name.Local]; ok {
						if aa, ok := a.([]interface{}); ok {
							aa = append(aa, cMap)
							mapStack[len(mapStack)-1][se.Name.Local] = aa
						} else {
							if am, ok := a.(map[string]interface{}); ok {
								aa := []interface{}{am, cMap}
								mapStack[len(mapStack)-1][se.Name.Local] = aa
							} else {
								log.Printf("Typing Error: %T", a)
							}
						}
					} else {
						cMap["_contents"] = string(curString)
						mapStack[len(mapStack)-1][se.Name.Local] = cMap
					}
				} else {
					//the child structure has no substructures, so we'll be treating it like string
					if a, ok := mapStack[len(mapStack)-1][se.Name.Local]; ok {
						if aa, ok := a.([]string); ok {
							aa = append(aa, string(curString))
							mapStack[len(mapStack)-1][se.Name.Local] = aa
						} else {
							if as, ok := a.(string); ok {
								aa := []string{as, string(curString)}
								mapStack[len(mapStack)-1][se.Name.Local] = aa
							} else {
								log.Printf("Typing Error %T", a)
							}
						}
					} else {
						mapStack[len(mapStack)-1][se.Name.Local] = string(curString)
					}
				}
			}

			if len(mapStack) == 1 {
				c := mapStack[0]
				if len(c) > 0 {
					t := map[string]interface{}{}
					for k := range c {
						t[k] = c[k]
					}
					out <- t
				}
				mapStack[0] = map[string]interface{}{}
			}

		case xml.CharData:
			curString = append(curString, se...)
			//default:
			//	log.Printf("Unknown Element: %#v\n", se)
		}
	}
}

func (ml *XMLLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	//log.Printf("Starting XML Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetConfig(), nil)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return nil, fmt.Errorf("File Not Found: %s", input)
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

	procChan := make(chan map[string]interface{}, 100)
	go func() {
		jStr, err := xj.Convert(hd)
		if err == nil {
			data := map[string]any{}
			if err = json.Unmarshal(jStr.Bytes(), &data); err == nil {
				procChan <- data
			}
		}
		//log.Printf("Starting XML Read")
		//xmlStream(hd, procChan)
		//log.Printf("Yes Done")
		//log.Printf("Done Loading")
		close(procChan)
	}()
	return procChan, nil
}

func (ml *XMLLoadStep) GetConfigFields() []config.ConfigVar {
	out := []config.ConfigVar{}
	for _, s := range evaluate.ExpressionIDs(ml.Input) {
		out = append(out, config.ConfigVar{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}

package extractors

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	"github.com/bmeg/sifter/transform"
)

type XMLLoadStep struct {
	Input         string         `json:"input"`
	Transform     transform.Pipe `json:"transform"`
	SkipIfMissing bool           `json:"skipIfMissing"`
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
								log.Printf("Typing Error")
							}
						}
					} else {
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
								log.Printf("Typing Error")
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
		default:
			log.Printf("Unknown Element: %#v\n", se)
		}
	}
}

func (ml *XMLLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	log.Printf("Starting XML Load")
	input, err := evaluate.ExpressionString(ml.Input, task.GetInputs(), nil)
	inputPath, err := task.AbsPath(input)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)

	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	procChan := make(chan map[string]interface{}, 100)
	go func() {
		log.Printf("Starting XML Read")
		xmlStream(file, procChan)
		log.Printf("Yes Done")
		log.Printf("Done Loading")
		close(procChan)
	}()
	return procChan, nil
}

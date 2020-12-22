package graphbuild

import (
	//"io"
	"os"
	//"strings"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/evaluate"
	"github.com/ghodss/yaml"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/grip/protoutil"
)

type Mapping struct {
	Domains map[string]*DomainMap `json:"domains"`
}

type DomainMap map[string]*ObjectMap

type TableLookupTransform struct {
	Table string `json:"table"`
	From  string `json:"From"`
}

type FieldTransform struct {
	Template    string                `json:"template"`
	TableLookup *TableLookupTransform `json:"tableLookup"`
	table       map[string]string
	field       string
}

type EdgeTransform struct {
	Template string `json:"template"`
}

type ObjectMap struct {
	Fields map[string]*FieldTransform `json:"fields"`
	Edges  map[string]*EdgeTransform  `json:"edges"`
}

func LoadMapping(path string, inputDir string) (*Mapping, error) {
	o := Mapping{}
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read data at path %s: \n%v", path, err)
	}
	if err := yaml.Unmarshal(raw, &o); err != nil {
		return nil, fmt.Errorf("failed to load graph mapping %s : %s", path, err)
	}

	for _, domain := range o.Domains {
		for _, cls := range *domain {
			for f, field := range cls.Fields {
				field.Init(f, inputDir)
			}
		}
	}

	return &o, nil
}

func (o *ObjectMap) MapObject(d map[string]interface{}) map[string]interface{} {
	if i, ok := o.Fields["_gid"]; ok {
		sid, err := evaluate.ExpressionString(i.Template, nil, d)
		if err == nil {
			d["id"] = sid
		}
	}
	for _, f := range o.Fields {
		d = f.Run(d)
	}
	return d
}

func (f *FieldTransform) Init(field string, inputDir string) error {
	f.field = field
	if f.TableLookup != nil {
		f.table = map[string]string{}

		p := filepath.Join(inputDir, fmt.Sprintf("%s.table.gz", f.TableLookup.Table))

		fhd, err := os.Open(p)
		if err != nil {
			log.Printf("Error Opening Table: %s", err)
			return err
		}
		defer fhd.Close()
		log.Printf("Reading Table File %s", p)
		hd, err := gzip.NewReader(fhd)
		if err != nil {
			return err
		}

		r, err := golib.ReadLines(hd)
		if err != nil {
			return err
		}
		parse := golib.CSVReader{}
		parse.Comma = "\t"
		var header []string
		for row := range parse.Read(r) {
			if header == nil {
				header = row
			} else {
				if len(row) == 2 {
					f.table[row[0]] = row[1]
				}
			}
		}
	}
	return nil
}

func (f *FieldTransform) Run(d map[string]interface{}) map[string]interface{} {
	if f.table != nil {
		if i, ok := d[f.TableLookup.From]; ok {
			if iString, ok := i.(string); ok {
				if o, ok := f.table[iString]; ok {
					//log.Printf("Translate %s to %s", iString, o)
					d[f.field] = o
				} else {
					log.Printf("Missing from %s translation table: %s", f.TableLookup.From, iString)
				}
			}
		} else {
			log.Printf("Field Missing: %s", f.TableLookup.From)
		}
	}
	if f.Template != "" {
		val, err := evaluate.ExpressionString(f.Template, nil, d)
		if err == nil {
			d[f.field] = val
		}
	}
	return d
}

func (et *EdgeTransform) Run(e *gripql.Edge) *gripql.Edge {
	if et.Template != "" {
		d := protoutil.AsMap(e.Data)
		if d == nil {
			d = map[string]interface{}{}
		}
		d["_to"] = e.To
		d["_from"] = e.From
		d["_label"] = e.Label
		val, err := evaluate.ExpressionString(et.Template, nil, d)
		if err == nil {
			e.To = val
		}
	}
	return e
}
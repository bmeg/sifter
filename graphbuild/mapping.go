package graphbuild

import (
	//"io"
	"os"
	"strings"
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
	AllVertex   *VertexFieldMapping         `json:"allVertex"`
	AllEdge     *EdgeFieldMapping           `json:"allEdge"`
	Domains map[string]*DomainMap           `json:"domains"`
}

type VertexFieldMapping struct {
	Fields map[string]*FieldTransform `json:"fields"`
}

type EdgeFieldMapping struct {
	Fields map[string]*FieldTransform `json:"fields"`
}

type DomainMap map[string]*VertexTransform

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
	EdgeFieldMapping
	DomainFilter  bool                `json:"domainFilter"`
	ToDomain string                   `json:"toDomain"`
	Sep    *string                    `json:"sep"`
}

type VertexTransform struct {
	VertexFieldMapping
	IdField   string                  `json:"idField"`
	Domain string                     `json:"domain"`
	Sep    *string                    `json:"sep"`
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
			for _, edge := range cls.Edges {
				for f, field := range edge.Fields {
					field.Init(f, inputDir)
				}
			}
		}
	}

	if o.AllVertex != nil {
		for f, field := range o.AllVertex.Fields {
			field.Init(f, inputDir)
		}
		if o.AllEdge != nil {
			for f, field := range o.AllEdge.Fields {
				field.Init(f, inputDir)
			}
		}
	}

	return &o, nil
}


func (m *Mapping) GetVertexDomains() []string {
	out := []string{}
	for _, d := range m.Domains {
		for _, v := range *d {
			out = append(out, v.Domain)
		}
	}
	return out
}

func (m *Mapping) GetEdgeEndDomains() [][]string {
	out := [][]string{}
	for _, d := range m.Domains {
		for _, v := range *d {
			for _, e := range v.Edges {
				out = append(out, []string{v.Domain,e.ToDomain})
				out = append(out, []string{e.ToDomain,v.Domain})
			}
		}
	}
	return out
}

func (vt *VertexTransform) Run(v *gripql.Vertex) *gripql.Vertex {
	o := vt.VertexFieldMapping.Run(v)
	if vt.Domain != "" && !strings.HasPrefix(o.Gid, vt.Domain) {
		sep := ":"
		if vt.Sep != nil {
			sep = *vt.Sep
		}
		o.Gid = vt.Domain + sep + o.Gid
	}
	return o
}

func (vt *VertexTransform) VertexObjectFix(obj map[string]interface{}) map[string]interface{} {
	if vt.IdField != "" {
		if g, ok := obj[vt.IdField]; ok {
			if gStr, ok := g.(string); ok {
				obj["id"] = gStr
				obj["_gid"] = gStr
			}
		}
	}
	if gid, ok := obj["_gid"]; ok {
		if gStr, ok := gid.(string); ok {
			if vt.Domain != "" && !strings.HasPrefix(gStr, vt.Domain) {
				sep := ":"
				if vt.Sep != nil {
					sep = *vt.Sep
				}
				obj["_gid"] = vt.Domain + sep + gStr
				obj["id"] = vt.Domain + sep + gStr
			}
		}
	}
	return obj
}

func (vfm *VertexFieldMapping) Run(v *gripql.Vertex) *gripql.Vertex {
	d := protoutil.AsMap(v.Data)
	if d == nil {
		d = map[string]interface{}{}
	}
	d["_gid"] = v.Gid
	d["_label"] = v.Label

	for _, f := range vfm.Fields {
		d = f.Run(d)
	}
	gid := ""
	if g, ok := d["_gid"]; ok {
		gid = g.(string)
	}
	o := gripql.Vertex{Gid: gid, Label:d["_label"].(string)}
	delete(d, "_gid")
	delete(d, "_label")
	o.Data = protoutil.AsStruct(d)
	return &o
}

func (et *EdgeTransform) Run(e *gripql.Edge) *gripql.Edge {
	o := et.EdgeFieldMapping.Run(e)
	if et.ToDomain != "" && !strings.HasPrefix(o.To, et.ToDomain) {
		sep := ":"
		if et.Sep != nil {
			sep = *et.Sep
		}
		if et.DomainFilter {
			o.To = "" //domain filter is on, but dest edge doesn't match, so set to "", so it's filtered out later
		} else {
			o.To = et.ToDomain + sep + o.To
		}
	}
	return o
}

func (et *EdgeFieldMapping) Run(e *gripql.Edge) *gripql.Edge {
	d := protoutil.AsMap(e.Data)
	if d == nil {
		d = map[string]interface{}{}
	}
	d["_to"] = e.To
	d["_from"] = e.From
	d["_label"] = e.Label

	for _, f := range et.Fields {
		d = f.Run(d)
	}
	gid := ""
	if g, ok := d["_gid"]; ok {
		gid = g.(string)}
	o := gripql.Edge{Gid: gid, From: d["_from"].(string), To:d["_to"].(string), Label:d["_label"].(string)}
	delete(d, "_gid")
	delete(d, "_to")
	delete(d, "_from")
	delete(d, "_label")
	o.Data = protoutil.AsStruct(d)
	return &o
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

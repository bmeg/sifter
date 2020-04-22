package graph

import (
  "os"
  "log"
  "fmt"
  "path/filepath"
	"github.com/bmeg/sifter/schema"
  "github.com/bmeg/grip/gripql"

)

type DomainClassInfo struct {
  emitter GraphEmitter
  gc      *GraphCheck
  om      *ObjectMap
  vertCount int64
  edgeCount int64
}

type DomainInfo struct {
  emitter GraphEmitter
  gc      *GraphCheck
  dm      *DomainMap
  classes map[string]*DomainClassInfo
}

type Builder struct {
	emitter GraphEmitter
	sc      schema.Schemas
	gm      *GraphMapping
  gc      *GraphCheck
  domains map[string]*DomainInfo
}

func NewBuilder(driver string, sc schema.Schemas, workdir string) (*Builder, error) {
	emitter, err := NewGraphEmitter(driver)
	if err != nil {
		return nil, err
	}
  gc, err := NewGraphCheck(workdir)
  if err != nil {
		return nil, err
	}
	return &Builder{sc: sc, emitter: emitter, domains:map[string]*DomainInfo{}, gc:gc}, nil
}

func (b *Builder) Close() {
  b.emitter.Close()
}

func (b *Builder) AddMapping(m *GraphMapping) {
	b.gm = m
}


func (b *Builder) GetDomain(prefix string) *DomainInfo {
  if x, ok := b.domains[prefix]; ok {
    return x
  }
  o := DomainInfo{emitter:b.emitter, classes:map[string]*DomainClassInfo{}, gc:b.gc}
  if x, ok := b.gm.Domains[prefix]; ok {
    o.dm = x
  }
  b.domains[prefix] = &o
  return &o
}

func (b *Builder) Process(prefix string, class string, in chan map[string]interface{}) {
	var m *ObjectMap
	if b.gm != nil {
		if x, ok := b.gm.Domains[prefix]; ok {
			if y, ok := (*x)[class]; ok {
				log.Printf("Using mapping: %s %s", prefix, class)
				m = y
			}
		}
	}

  d := b.GetDomain(prefix)
  c := d.GetClass(class)

	for obj := range in {
		if m != nil {
			obj = m.MapObject(obj)
		}
		err := GenerateGraph(&b.sc, class, obj, c)
		if err != nil {
			log.Printf("Graph Generation Error: %s.%s : %s", prefix, class, err)
		}
	}
}

func (b *Builder) Report(outdir string) {
  rout, err := os.Create(filepath.Join(outdir, "report.txt"))
  if err != nil {
    return
  }
  defer rout.Close()
  for d, i := range b.domains {
    fmt.Fprintf(rout, "Domain: %s\n", d)
    for c, j := range i.classes {
      fmt.Fprintf(rout, "\t%s\t%d\t%d\n", c, j.vertCount, j.edgeCount)
    }
  }
  mout, err := os.Create(filepath.Join(outdir, "missing.txt"))
  if err != nil {
    return
  }
  defer mout.Close()
  for v := range b.gc.GetEdgeVertices() {
      if ! b.gc.HasVertex(v) {
        fmt.Fprintf(mout, "%s (from %s)\n", v, b.gc.GetEdgeSource(v))
      }
  }
}

func (d *DomainInfo) GetClass(cls string) *DomainClassInfo {
  if x, ok := d.classes[cls]; ok {
    return x
  }
  o := DomainClassInfo{emitter:d.emitter, gc:d.gc}
  if x, ok := (*d.dm)[cls]; ok {
    o.om = x
  }
  d.classes[cls] = &o
  return &o
}


func (dc *DomainClassInfo) Close() {
  dc.emitter.Close()
}


func (dc *DomainClassInfo) EmitVertex(v *gripql.Vertex) error {
  dc.vertCount += 1
  if dc.om != nil {
    if l, ok := dc.om.Fields["_label"]; ok {
      v.Label = l.Template
    }
  }
  dc.gc.AddVertex(v.Gid)
  return dc.emitter.EmitVertex(v)
}

func (dc *DomainClassInfo) EmitEdge(e *gripql.Edge) error {
  dc.edgeCount += 1
  dc.gc.AddEdge(e.From, e.To)
  return dc.emitter.EmitEdge(e)
}

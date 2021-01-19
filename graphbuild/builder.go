package graphbuild

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/schema"
)

type DomainClassInfo struct {
	emitter   loader.GraphEmitter
	gc        *Check
	om        *VertexTransform
	vertCount int64
	edgeCount int64
}

type DomainInfo struct {
	emitter loader.GraphEmitter
	gc      *Check
	dm      *DomainMap
	classes map[string]*DomainClassInfo
}

type Builder struct {
	loader  loader.Loader
	emitter loader.GraphEmitter
	sc      schema.Schemas
	gm      *Mapping
	gc      *Check
	domains map[string]*DomainInfo
}

func NewBuilder(ld loader.Loader, sc schema.Schemas, workdir string) (*Builder, error) {
	emitter, err := ld.NewGraphEmitter()
	if err != nil {
		return nil, err
	}
	gc, err := NewGraphCheck(workdir)
	if err != nil {
		return nil, err
	}
	return &Builder{loader: ld, sc: sc, emitter: emitter, domains: map[string]*DomainInfo{}, gc: gc}, nil
}

func (b *Builder) Close() {
	b.loader.Close()
}

func (b *Builder) AddMapping(m *Mapping) {
	b.gm = m
}

func (b *Builder) GetDomain(prefix string) *DomainInfo {
	if x, ok := b.domains[prefix]; ok {
		return x
	}
	o := DomainInfo{emitter: b.emitter, classes: map[string]*DomainClassInfo{}, gc: b.gc}
	if x, ok := b.gm.Domains[prefix]; ok {
		o.dm = x
	} else {
		log.Printf("Domain info for %s not found", prefix)
		return nil
	}
	b.domains[prefix] = &o
	return &o
}

func (b *Builder) HasDomain(prefix string, class string) bool {
	d := b.GetDomain(prefix)
	if d == nil {
		return false
	}
	c := d.GetClass(class)
	if c == nil {
		return false
	}
	return true
}

func (b *Builder) Process(prefix string, class string, in chan map[string]interface{}) {
	var m *VertexTransform
	if b.gm != nil {
		if x, ok := b.gm.Domains[prefix]; ok {
			if y, ok := (*x)[class]; ok {
				log.Printf("Using mapping: %s %s", prefix, class)
				m = y
			}
		}
	}

	d := b.GetDomain(prefix)
	if d == nil {
		for range in {}
		return
	}
	c := d.GetClass(class)
	if c == nil {
		for range in {}
		return
	}

	for obj := range in {
		obj = c.om.VertexObjectFix(obj)
		err := b.GenerateGraph(m, class, obj, c)
		if err != nil {
			log.Printf("Graph Generation Error: %s.%s : %s", prefix, class, err)
		}
	}
}

func (b *Builder) GenerateGraph(vertMap *VertexTransform, class string, data map[string]interface{}, emitter loader.GraphEmitter) error {
	if o, err := b.sc.Generate(class, data); err == nil {
		for _, j := range o {
			if j.Vertex != nil {
				if b.gm.AllVertex != nil {
					j.Vertex = b.gm.AllVertex.Run(j.Vertex)
				}
				if vertMap != nil {
					j.Vertex = vertMap.Run(j.Vertex)
				}
				emitter.EmitVertex(j.Vertex)
			} else if j.OutEdge != nil || j.InEdge != nil {
				var edge *gripql.Edge
				if j.OutEdge != nil {
					edge = j.OutEdge
				}
				if j.InEdge != nil {
					edge = j.InEdge
				}
				if b.gm.AllEdge != nil {
					edge = b.gm.AllEdge.Run(edge)
				}
				if vertMap != nil {
					if em, ok := vertMap.Edges[edge.Label]; ok {
						edge = em.Run(edge)
					} else if j.OutEdge != nil {
						//report if an outgoing edge does have mapping information
						log.Printf("Mapping for edge %s not found", edge.Label )
					}
				}
				if edge.To != "" && edge.From != "" && edge.Label != "" {
					emitter.EmitEdge(edge)
				}
			}
		}
	} else {
		return err
	}
	return nil
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
		if !b.gc.HasVertex(v) {
			fmt.Fprintf(mout, "%s (from %s)\n", v, b.gc.GetEdgeSource(v))
		}
	}
}

func (d *DomainInfo) GetClass(cls string) *DomainClassInfo {
	if x, ok := d.classes[cls]; ok {
		return x
	}
	o := DomainClassInfo{emitter: d.emitter, gc: d.gc}
	if x, ok := (*d.dm)[cls]; ok {
		o.om = x
	}
	d.classes[cls] = &o
	return &o
}

func (dc *DomainClassInfo) EmitVertex(v *gripql.Vertex) error {
	dc.vertCount++
	if dc.om != nil {
		if l, ok := dc.om.Fields["_label"]; ok {
			v.Label = l.Template
		}
	}
	dc.gc.AddVertex(v.Gid)
	return dc.emitter.EmitVertex(v)
}

func (dc *DomainClassInfo) EmitEdge(e *gripql.Edge) error {
	dc.edgeCount++
	dc.gc.AddEdge(e.From, e.To)
	return dc.emitter.EmitEdge(e)
}

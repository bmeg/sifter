package graph

import (
	"github.com/bmeg/sifter/schema"
	"log"
)

type Builder struct {
	emitter GraphEmitter
	sc      schema.Schemas
	gm      *GraphMapping
}

func NewBuilder(driver string, sc schema.Schemas) (*Builder, error) {
	emitter, err := NewGraphEmitter(driver)
	if err != nil {
		return nil, err
	}
	return &Builder{sc: sc, emitter: emitter}, nil
}

func (b *Builder) AddMapping(m *GraphMapping) {
	b.gm = m
}

func (b *Builder) Process(prefix string, class string, in chan map[string]interface{}) {
	var m *ObjectMap
	if b.gm != nil {
		if x, ok := b.gm.Domains[prefix]; ok {
			if y, ok := x[class]; ok {
				log.Printf("Using mapping: %s %s", prefix, class)
				m = &y
			}
		}
	}
	for obj := range in {
		if m != nil {
			obj = m.MapObject(obj)
		}
		err := GenerateGraph(&b.sc, class, obj, b.emitter)
		if err != nil {
			log.Printf("Graph Generation Error: %s.%s : %s", prefix, class, err)
		}
	}
}

package graph

import (
  "log"
  "github.com/bmeg/sifter/schema"
)

type Builder struct {
  emitter GraphEmitter
  sc schema.Schemas
}


func NewBuilder(driver string, sc schema.Schemas) (*Builder, error) {
  emitter, err := NewGraphEmitter(driver)
  if err != nil {
    return nil, err
  }
  return &Builder{sc:sc, emitter:emitter}, nil
}


func (b *Builder) Process(prefix string, class string, in chan map[string]interface{} ) {
  for obj := range in {
      err := GenerateGraph(&b.sc, class, obj, b.emitter)
      if err != nil {
        log.Printf("Graph Generation Error: %s", err)
      }
  }
}

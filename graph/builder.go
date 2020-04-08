package graph

import (
  "github.com/bmeg/sifter/schema"
)

type Builder struct {
  sc schema.Schemas
}


func NewBuilder(sc schema.Schemas) (*Builder, error) {
  return &Builder{sc:sc}, nil
}




func (b *Builder) Process(prefix string, class string, in chan map[string]interface{} ) {
  for range in {

  }
}

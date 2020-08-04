package emitter

import (
	"fmt"
	"net/url"
	"github.com/bmeg/sifter/schema"
)

type TableEmitter interface {
	EmitRow(values map[string]interface{}) error
	Close()
}

type Emitter interface {
	EmitObject(prefix string, objClass string, e map[string]interface{}) error
	EmitTable(prefix string, columns []string, sep rune) TableEmitter
	Close()
}


func NewEmitter(driver string, sc *schema.Schemas) (Emitter, error) {
	u, _ := url.Parse(driver)
	if u.Scheme == "stdout" {
		return StdoutEmitter{schemas:sc}, nil
	}
	if u.Scheme == "dir" {
		return NewDirEmitter( u.Host + u.Path, sc ), nil
	}
	return nil, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

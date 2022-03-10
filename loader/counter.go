package loader

import (
	"sync/atomic"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"
)

type CountLoader struct {
	l        Loader
	rate     uint64
	callback func(uint64)
	count    uint64
}

type CountDataEmitter struct {
	d  DataEmitter
	cl *CountLoader
}

type CountGraphEmitter struct {
	g  GraphEmitter
	cl *CountLoader
}

func NewLoadCounter(l Loader, rate uint64, callback func(uint64)) Loader {
	return &CountLoader{l: l, rate: rate, callback: callback}
}

func (cl *CountLoader) Close() {
	cl.l.Close()
}

func (cl *CountLoader) NewDataEmitter(sch *schema.Schemas) (DataEmitter, error) {
	o, err := cl.l.NewDataEmitter(sch)
	return &CountDataEmitter{o, cl}, err
}

func (cl *CountLoader) NewGraphEmitter() (GraphEmitter, error) {
	o, err := cl.l.NewGraphEmitter()
	return &CountGraphEmitter{o, cl}, err
}

func (cl *CountLoader) increment() {
	v := atomic.AddUint64(&cl.count, 1)
	if (v % cl.rate) == 0 {
		cl.callback(v)
	}
}

func (cd *CountDataEmitter) Close() {
	cd.d.Close()
}

func (cd *CountDataEmitter) Emit(name string, e map[string]interface{}) error {
	cd.cl.increment()
	return cd.d.Emit(name, e)
}

func (cd *CountDataEmitter) EmitObject(prefix string, objClass string, e map[string]interface{}) error {
	cd.cl.increment()
	return cd.d.EmitObject(prefix, objClass, e)
}

func (cg *CountGraphEmitter) EmitVertex(v *gripql.Vertex) error {
	cg.cl.increment()
	return cg.g.EmitVertex(v)
}

func (cg *CountGraphEmitter) EmitEdge(e *gripql.Edge) error {
	cg.cl.increment()
	return cg.g.EmitEdge(e)
}

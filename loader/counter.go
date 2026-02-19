package loader

import (
	"sync/atomic"
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
	cl *CountLoader
}

func NewLoadCounter(l Loader, rate uint64, callback func(uint64)) Loader {
	return &CountLoader{l: l, rate: rate, callback: callback}
}

func (cl *CountLoader) Close() {
	cl.l.Close()
}

func (cl *CountLoader) NewDataEmitter() (DataEmitter, error) {
	o, err := cl.l.NewDataEmitter()
	return &CountDataEmitter{o, cl}, err
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

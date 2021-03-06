package template

import (
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/schema"
)

type setupLoader func(opts map[string]string) (loader.Loader, error)

var LoadTemplates = map[string]setupLoader{
	"dir": func(opts map[string]string) (loader.Loader, error) {
		return loader.NewLoader("dir://./")
	},
	"grip": func(opts map[string]string) (loader.Loader, error) {
		grip := "grip://localhost:8202/sifter"
		if g, ok := opts["grip"]; ok {
			grip = g
		}
		ld, err := loader.NewLoader(grip)
		if err != nil {
			return nil, err
		}
		return WrappedGraphLoader{ld}, nil
	},
}

type WrappedGraphLoader struct {
	ld loader.Loader
}

func (wg WrappedGraphLoader) NewGraphEmitter() (loader.GraphEmitter, error) {
	return wg.ld.NewGraphEmitter()
}

func (wg WrappedGraphLoader) NewDataEmitter(sc *schema.Schemas) (loader.DataEmitter, error) {
	em, err := wg.ld.NewGraphEmitter()
	if err != nil {
		return nil, err
	}
	dl := loader.GraphTransformer(em, sc)
	return dl, nil
}

func (wg WrappedGraphLoader) Close() {
	wg.ld.Close()
}

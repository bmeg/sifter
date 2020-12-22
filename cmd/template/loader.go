package template

import (
  "github.com/bmeg/sifter/loader"
)

type setupLoader func(opts map[string]string) (loader.Loader, error)

var LoadTemplates = map[string]setupLoader{
  "dir" : func(opts map[string]string) (loader.Loader, error) {
    return loader.NewLoader("dir://./")
  },
}

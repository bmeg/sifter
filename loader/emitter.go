package loader

import (
	"fmt"
	"net/url"
)

type DataEmitter interface {
	Emit(name string, e map[string]interface{}, useName bool) error
	Close()
}

type Loader interface {
	NewDataEmitter() (DataEmitter, error)
	Close()
}

func NewLoader(driver string) (Loader, error) {
	u, _ := url.Parse(driver)
	if u.Scheme == "stdout" {
		return StdoutLoader{}, nil
	}
	if u.Scheme == "dir" {
		return NewDirLoader(u.Host + u.Path), nil
	}
	return nil, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

func GraphExists(server string, graph string, args string) (bool, error) {
	u, _ := url.Parse(server)

	if u.Scheme == "stdout" {
		return false, nil
	}
	if u.Scheme == "dir" {
		return false, nil
	}
	return false, fmt.Errorf("Unknown driver: %s", u.Scheme)
}

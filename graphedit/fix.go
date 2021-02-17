package graphedit

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/bmeg/sifter/extractors"
	"github.com/bmeg/sifter/loader"

	"google.golang.org/protobuf/types/known/structpb"
)

type Config struct {
	Domains map[string]*DomainConfig `json:"domains"`
}

type DomainConfig struct {
	IDMap         string `json:"idMap"`
	StoreOriginal string `json:"storeOriginal"`
	mapping       map[string]string
}

func LoadGraphEdit(path string) (*Config, error) {
	path, _ = filepath.Abs(path)
	baseDir := filepath.Dir(path)

	o := Config{}
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read data at path %s: \n%v", path, err)
	}
	if err := yaml.Unmarshal(raw, &o); err != nil {
		return nil, fmt.Errorf("failed to load graph mapping %s : %s", path, err)
	}

	for _, domain := range o.Domains {
		if domain.IDMap != "" {
			idmap := map[string]string{}
			path := filepath.Join(baseDir, domain.IDMap)
			df, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			scanner := bufio.NewScanner(df)
			buf := make([]byte, 64*1042*1024)
			scanner.Buffer(buf, len(buf))
			for scanner.Scan() {
				tmp := strings.Split(scanner.Text(), "\t")
				idmap[tmp[0]] = tmp[1]
			}
			domain.mapping = idmap
		}
	}
	return &o, nil
}

func (conf *Config) EditVertexFile(srcPath, dstPath string) error {
	stream, err := extractors.LoadVertexFile(srcPath)
	if err != nil {
		log.Printf("File Error: %s", err)
		return err
	}

	out, err := loader.NewBGZipGraphEmitter(dstPath)
	if err != nil {
		log.Printf("File Error: %s", err)
		return err
	}
	for e := range stream {
		out.EmitVertex(e)
	}
	out.Close()
	return nil
}

func (conf *Config) EditEdgeFile(srcPath, dstPath string) error {

	stream, err := extractors.LoadEdgeFile(srcPath)
	if err != nil {
		log.Printf("File Error: %s", err)
		return err
	}

	out, err := loader.NewBGZipGraphEmitter(dstPath)
	if err != nil {
		log.Printf("File Error: %s", err)
		return err
	}

	for e := range stream {
		toFound := false
		fromFound := false
		for d, dinfo := range conf.Domains {
			if !toFound && strings.HasPrefix(e.To, d+":") {
				toID := e.To[len(d)+1 : len(e.To)]
				if newID, ok := dinfo.mapping[toID]; ok {
					e.To = d + ":" + newID
					toFound = true
					if dinfo.StoreOriginal != "" {
						if e.Data == nil {
							e.Data = &structpb.Struct{Fields: map[string]*structpb.Value{}}
						}
						e.Data.Fields[dinfo.StoreOriginal], _ = structpb.NewValue(toID)
					}
				}
			}
			if !fromFound && strings.HasPrefix(e.From, d+":") {
				fromID := e.From[len(d)+1 : len(e.From)]
				if newID, ok := dinfo.mapping[fromID]; ok {
					e.From = d + ":" + newID
					fromFound = true
					if dinfo.StoreOriginal != "" {
						e.Data.Fields[dinfo.StoreOriginal], _ = structpb.NewValue(fromID)
					}
				}
			}
		}
		out.EmitEdge(e)
	}
	out.Close()
	return nil
}

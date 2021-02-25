package graphedit

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/bmeg/grip/gripql"

	"github.com/bmeg/sifter/extractors"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/transform"

	"google.golang.org/protobuf/types/known/structpb"
)

type Config struct {
	RuleMap []RuleMapping    `json:ruleMap`
	Rules   map[string]*Rule `json:rules`
}

type RuleMapping struct {
	Path       string `json:"path"`
	ToPrefix   string `json:"toPrefix"`
	FromPrefix string `json:"fromPrefix"`
	Rule       string `json:"rule"`
}

type Rule struct {
	IgnorePrefix  string         `json:"ignorePrefix"`
	MissingPrefix string         `json:"missingPrefix"`
	ToIDMap       string         `json:"toIDMap"`
	FromIDMap     string         `json:"fromIDMap"`
	StoreOriginal string         `json:"storeOriginal"`
	Transform     transform.Pipe `json:"transform"`
	Omit          bool           `json:"omit"`
	toMapping     map[string]string
	fromMapping   map[string]string
}

func (r *Rule) FixVertex(v *gripql.Vertex, out loader.GraphEmitter) {
	out.EmitVertex(v)
}

func (r *Rule) FixEdge(e *gripql.Edge, out loader.GraphEmitter) {
	if r.toMapping != nil {
		toID := e.To
		if len(r.IgnorePrefix) > 0 {
			toID = e.To[len(r.IgnorePrefix):len(e.To)]
		}
		if newID, ok := r.toMapping[toID]; ok {
			e.To = r.IgnorePrefix + newID
			if r.StoreOriginal != "" {
				if e.Data == nil {
					e.Data = &structpb.Struct{Fields: map[string]*structpb.Value{}}
				}
				e.Data.Fields[r.StoreOriginal], _ = structpb.NewValue(toID)
			}
		} else if r.MissingPrefix != "" {
			e.To = r.MissingPrefix + toID
		}
	} else if r.fromMapping != nil {
		fromID := e.From
		if len(r.IgnorePrefix) > 0 {
			fromID = e.From[len(r.IgnorePrefix):len(e.From)]
		}
		if newID, ok := r.fromMapping[fromID]; ok {
			e.From = r.IgnorePrefix + newID
			if r.StoreOriginal != "" {
				if e.Data == nil {
					e.Data = &structpb.Struct{Fields: map[string]*structpb.Value{}}
				}
				e.Data.Fields[r.StoreOriginal], _ = structpb.NewValue(fromID)
			}
		} else if r.MissingPrefix != "" {
			e.From = r.MissingPrefix + fromID
		}
	}
	out.EmitEdge(e)
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

	for _, domain := range o.Rules {
		if domain.ToIDMap != "" {
			idmap := map[string]string{}
			path := filepath.Join(baseDir, domain.ToIDMap)
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
			domain.toMapping = idmap
		}
		if domain.FromIDMap != "" {
			idmap := map[string]string{}
			path := filepath.Join(baseDir, domain.FromIDMap)
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
			domain.fromMapping = idmap
		}
	}
	return &o, nil
}

func (conf *Config) EditVertexFile(srcPath, dstPath string) error {
	for _, rm := range conf.RuleMap {
		if rm.Path != "" {
			log.Printf("Checking: %s %s", rm.Path, srcPath)
			if matched, _ := regexp.Match(rm.Path, []byte(srcPath)); matched {
				if r, ok := conf.Rules[rm.Rule]; ok {
					if r.Omit {
						log.Printf("Skipping: %s", srcPath)
						return nil
					}
				}
			}
		}
	}

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
	for v := range stream {
		//no vertex editing yet
		out.EmitVertex(v)
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
		ruleFound := false
		for _, rm := range conf.RuleMap {
			if rm.ToPrefix != "" && strings.HasPrefix(e.To, rm.ToPrefix) {
				ruleFound = true
				if r, ok := conf.Rules[rm.Rule]; ok {
					r.FixEdge(e, out)
				}
			} else if rm.FromPrefix != "" && strings.HasPrefix(e.From, rm.FromPrefix) {
				ruleFound = true
				if r, ok := conf.Rules[rm.Rule]; ok {
					r.FixEdge(e, out)
				}
			}
		}
		if !ruleFound {
			out.EmitEdge(e)
		}
	}
	out.Close()
	return nil
}

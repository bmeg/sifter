package graphmanifest

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func joinMapKeys(s map[string]bool) string {
	o := []string{}
	for k := range s {
		o = append(o, k)
	}
	return strings.Join(o, ",")
}

func mapKeys(s map[string]bool) []string {
	o := []string{}
	for k := range s {
		o = append(o, k)
	}
	return o
}

type FileInfo struct {
	FileType string   `json:"fileType"`
	Path     string   `json:"path"`
	Gid      *string  `json:"gid"`
	FromGid  *string  `json:"fromGid"`
	ToGid    *string  `json:"toGid"`
	Labels   []string `json:"labels"`
}

func fileScanner(path string) (FileInfo, error) {
	if strings.HasSuffix(path, ".Vertex.json.gz") {
		log.Printf("Scanning %s", path)
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return FileInfo{}, err
		}
		gr, _ := gzip.NewReader(fr)
		scanner := bufio.NewScanner(gr)
		bufferSize := 32 * 1024 * 1024
		buffer := make([]byte, bufferSize)
		scanner.Buffer(buffer, bufferSize)

		labels := map[string]bool{}
		gidSubStr := ""
		for scanner.Scan() {
			d := map[string]interface{}{}
			err := json.Unmarshal([]byte(scanner.Text()), &d)
			if err == nil {
				gid := d["gid"].(string)
				label := d["label"].(string)
				labels[label] = true
				if gidSubStr == "" {
					gidSubStr = gid
				} else {
					i := 0
					for ; i < len(gid) && i < len(gidSubStr) && gid[i] == gidSubStr[i]; i++ {
					}
					if i > 2 && i != len(gidSubStr) {
						gidSubStr = gid[:i]
					}
				}
			}
		}
		return FileInfo{FileType: "edge", Path: path, Gid: &gidSubStr, Labels: mapKeys(labels)}, nil
	} else if strings.HasSuffix(path, ".Edge.json.gz") {
		log.Printf("Scanning %s", path)
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return FileInfo{}, err
		}
		gr, _ := gzip.NewReader(fr)
		scanner := bufio.NewScanner(gr)
		bufferSize := 32 * 1024 * 1024
		buffer := make([]byte, bufferSize)
		scanner.Buffer(buffer, bufferSize)

		labels := map[string]bool{}
		toGidSubStr := ""
		fromGidSubStr := ""
		for scanner.Scan() {
			d := map[string]interface{}{}
			err := json.Unmarshal([]byte(scanner.Text()), &d)
			if err == nil {
				var toGid, fromGid, label string
				if t, ok := d["to"]; ok {
					if toGid, ok = t.(string); !ok {
						log.Printf("to is not string in %s", d)
					}
				} else {
					log.Printf("to not found in %s", d)
				}
				fromGid = d["from"].(string)
				label = d["label"].(string)
				labels[label] = true
				if toGidSubStr == "" {
					toGidSubStr = toGid
				} else {
					i := 0
					for ; i < len(toGid) && i < len(toGidSubStr) && toGid[i] == toGidSubStr[i]; i++ {
					}
					if i > 2 && i != len(toGidSubStr) {
						toGidSubStr = toGid[:i]
					}
				}
				if fromGidSubStr == "" {
					fromGidSubStr = fromGid
				} else {
					i := 0
					for ; i < len(fromGid) && i < len(fromGidSubStr) && fromGid[i] == fromGidSubStr[i]; i++ {
					}
					if i > 2 && i != len(fromGidSubStr) {
						fromGidSubStr = fromGid[:i]
					}
				}
			}
		}
		return FileInfo{FileType: "edge", Path: path, FromGid: &fromGidSubStr, ToGid: &toGidSubStr, Labels: mapKeys(labels)}, nil
	}
	return FileInfo{}, fmt.Errorf("Unknown file suffix: %s", path)
}

func ScanDir(path string, workerCount int) chan FileInfo {
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > 50 {
		workerCount = 50
	}

	out := make(chan FileInfo, workerCount)
	fileNames := make(chan string, workerCount)

	wg := &sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			for n := range fileNames {
				if o, err := fileScanner(n); err == nil {
					out <- o
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	go func() {
		defer close(fileNames)
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			fileNames <- path
			return nil
		})
	}()

	return out
}

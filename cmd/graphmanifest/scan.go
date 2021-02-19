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
)

func joinMapKeys(s map[string]bool) string {
	o := []string{}
	for k := range s {
		o = append(o, k)
	}
	return strings.Join(o, ",")
}

func ScanDir(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".Vertex.json.gz") {
			log.Printf("Scanning %s", path)
			fr, err := os.Open(path)
			defer fr.Close()
			if err != nil {
				return err
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
			fmt.Printf("%s\t%s\t%s\n", path, gidSubStr, joinMapKeys(labels))
		} else if strings.HasSuffix(path, ".Edge.json.gz") {
			log.Printf("Scanning %s", path)
			fr, err := os.Open(path)
			defer fr.Close()
			if err != nil {
				return err
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
					toGid := d["to"].(string)
					fromGid := d["from"].(string)
					label := d["label"].(string)
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
			fmt.Printf("%s\t%s\t%s\t%s\n", path, fromGidSubStr, toGidSubStr, joinMapKeys(labels))
		}
		return nil
	})
	return err
}

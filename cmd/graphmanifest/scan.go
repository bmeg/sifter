package graphmanifest

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
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
			//fmt.Printf("%s\n", path)
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
		}
		return nil
	})
	return err
}

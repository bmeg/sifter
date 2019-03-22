package manager

import (
	"regexp"

	"github.com/bmeg/golib"
	gripUtil "github.com/bmeg/grip/util"
)

type ManifestLoadStep struct {
	Input   string `json:"input"`
	BaseURL string `json:"baseURL"`
}

var vertexRE *regexp.Regexp = regexp.MustCompile(".Vertex.json")
var edgeRE *regexp.Regexp = regexp.MustCompile(".Edge.json")

func (ml *ManifestLoadStep) Run(man *Task) error {
	man.Printf("loading manifest %s", ml.Input)
	lines, err := golib.ReadFileLines(man.Path(ml.Input))
	if err != nil {
		return err
	}
	entries := [][]byte{}
	for l := range lines {
		if len(l) > 0 {
			entries = append(entries, l)
		}
	}

	for _, l := range entries {
		if vertexRE.Match(l) {
			url := ml.BaseURL + string(l)
			man.Printf("Download: %s", url)
			path, err := man.DownloadFile(url)
			if err != nil {
				man.Printf("Download Failure: %s %s", url, err)
			} else {
				man.Printf("Loading %s", path)
				for v := range gripUtil.StreamVerticesFromFile(path) {
					man.EmitVertex(v)
				}
			}
		}
	}

	for _, l := range entries {
		if edgeRE.Match(l) {
			url := ml.BaseURL + string(l)
			man.Printf("Download: %s", url)
			path, err := man.DownloadFile(url)
			if err != nil {
				man.Printf("Download Failure: %s %s", url, err)
			} else {
				man.Printf("Loading %s", path)
				for v := range gripUtil.StreamEdgesFromFile(path) {
					man.EmitEdge(v)
				}
			}
		}
	}

	return nil
}

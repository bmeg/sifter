package manager

import (
	"regexp"

	"github.com/bmeg/golib"
	gripUtil "github.com/bmeg/grip/util"
	"github.com/bmeg/sifter/evaluate"
)

type ManifestLoadStep struct {
	Input   string `json:"input"`
	BaseURL string `json:"baseURL"`
}

var vertexRE *regexp.Regexp = regexp.MustCompile(".Vertex.json")
var edgeRE *regexp.Regexp = regexp.MustCompile(".Edge.json")

func (ml *ManifestLoadStep) Run(task *Task) error {
	task.Printf("loading manifest %s", ml.Input)
	lines, err := golib.ReadFileLines(task.Path(ml.Input))
	if err != nil {
		return err
	}
	entries := [][]byte{}
	for l := range lines {
		if len(l) > 0 {
			entries = append(entries, l)
		}
	}

	baseURL, err := evaluate.ExpressionString(ml.BaseURL, task.Inputs)

	for _, l := range entries {
		if vertexRE.Match(l) {
			url := baseURL + string(l)
			task.Printf("Download: %s", url)
			path, err := task.DownloadFile(url, "")
			if err != nil {
				task.Printf("Download Failure: %s %s", url, err)
			} else {
				task.Printf("Loading %s", path)
				for v := range gripUtil.StreamVerticesFromFile(path) {
					task.EmitVertex(v)
				}
			}
		}
	}

	for _, l := range entries {
		if edgeRE.Match(l) {
			url := baseURL + string(l)
			task.Printf("Download: %s", url)
			path, err := task.DownloadFile(url, "")
			if err != nil {
				task.Printf("Download Failure: %s %s", url, err)
			} else {
				task.Printf("Loading %s", path)
				for v := range gripUtil.StreamEdgesFromFile(path) {
					task.EmitEdge(v)
				}
			}
		}
	}

	return nil
}

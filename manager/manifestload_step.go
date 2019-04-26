package manager

import (
	"os"
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

	mlInput, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
	if err != nil {
		task.Printf("Expression failed: %s", err)
		return err
	}

	task.Printf("loading manifest %s", mlInput)
	path, err := task.Path(mlInput)
	if err != nil {
		return err
	}
	lines, err := golib.ReadFileLines(path)
	if err != nil {
		task.Printf("Manifest failed to load: %s", err)
		return err
	}
	entries := [][]byte{}
	for l := range lines {
		if len(l) > 0 {
			entries = append(entries, l)
		}
	}

	baseURL, err := evaluate.ExpressionString(ml.BaseURL, task.Inputs, nil)

	task.Runtime.SetStepCountTotal(int64(len(entries)))
	for _, l := range entries {
		if vertexRE.Match(l) {
			url := baseURL + string(l)
			task.Printf("Download: %s", url)
			path, err := task.DownloadFile(url, "")
			if err != nil {
				task.Printf("Download Failure: %s %s", url, err)
			} else {
				task.Printf("Loading vertex file %s", path)
				for v := range gripUtil.StreamVerticesFromFile(path) {
					task.EmitVertex(v)
				}
				os.Remove(path)
			}
			task.Runtime.AddStepCount(1)
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
				task.Printf("Loading edge file %s", path)
				for v := range gripUtil.StreamEdgesFromFile(path) {
					task.EmitEdge(v)
				}
				os.Remove(path)
			}
			task.Runtime.AddStepCount(1)
		}
	}

	return nil
}

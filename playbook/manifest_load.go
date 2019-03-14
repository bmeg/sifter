package playbook

import (
  "log"
  "regexp"
  "github.com/bmeg/golib"
  "github.com/bmeg/grip/gripql"
  "github.com/bmeg/sifter/manager"
  gripUtil "github.com/bmeg/grip/util"
)

type ManifestLoadStep struct {
  Input string `json:"input"`
  BaseURL string `json:"baseURL"`
}

var vertexRE *regexp.Regexp = regexp.MustCompile(".Vertex.json")
var edgeRE *regexp.Regexp = regexp.MustCompile(".Edge.json")

func (ml *ManifestLoadStep) Load(man manager.Manager) chan gripql.GraphElement {
  log.Printf("loading manifest %s", ml.Input)
  out := make(chan gripql.GraphElement, 10)
  go func() {
    defer close(out)
    lines, err := golib.ReadFileLines(man.Path(ml.Input))
    if err != nil {
      return
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
        log.Printf("Download: %s", url)
        path, err := man.DownloadFile(url)
        if err != nil {
          log.Printf("Download Failure: %s %s", url, err)
        } else {
          log.Printf("Loading %s", path)
          for v := range gripUtil.StreamVerticesFromFile(path) {
            man.EmitVertex(v)
          }
        }
      }
    }

    for _, l := range entries {
      if edgeRE.Match(l) {
        url := ml.BaseURL + string(l)
        log.Printf("Download: %s", url)
        path, err := man.DownloadFile(url)
        if err != nil {
          log.Printf("Download Failure: %s %s", url, err)
        } else {
          log.Printf("Loading %s", path)
          for v := range gripUtil.StreamEdgesFromFile(path) {
            man.EmitEdge(v)
          }
        }
      }
    }

  }()

  return out
}

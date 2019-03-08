package playbook

import (
  "log"
  "github.com/bmeg/golib"
  "github.com/bmeg/grip/gripql"
  "github.com/bmeg/sifter/manager"
)

type ManifestLoadStep struct {
  Input string `json:"input"`
  BaseURL string `json:"baseURL"`
}

func (ml *ManifestLoadStep) Load(man manager.Manager) chan gripql.GraphElement {
  log.Printf("loading manifest %s", ml.Input)
  out := make(chan gripql.GraphElement, 10)
  go func() {
    defer close(out)
    lines, err := golib.ReadFileLines(man.Path(ml.Input))
    if err != nil {
      return
    }
    for l := range lines {
      if len(l) > 0 {
        url := ml.BaseURL + string(l)
        log.Printf("Download: %s", url)
        path, err := man.DownloadFile(url)
        if err != nil {
          log.Printf("Download Failure: %s %s", url, err)
        } else {
          log.Printf("Loading %s", path)
          elements, err := golib.ReadGzipLines(path)
          if err != nil {
            log.Printf("Error reading %s", err)
          } else {
            for elem := range(elements) {
              if len(elem) > 0 {
                log.Printf("%s", elem)
              }
            }
          }
        }
      }
    }
  }()

  return out
}

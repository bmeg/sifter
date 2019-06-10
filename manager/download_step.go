package manager

import (
	"log"

	"github.com/bmeg/sifter/evaluate"
)

type DownloadStep struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
	Output string `json:"output"`
}

func (ps *DownloadStep) Run(task *Task) error {
	srcURL, err := evaluate.ExpressionString(ps.Source, task.Inputs, nil)
	if err != nil {
		log.Printf("Expression failed: %s", err)
		return err
	}
	dstPath, err := evaluate.ExpressionString(ps.Dest, task.Inputs, nil)
	task.Printf("Downloading: %s to %s", srcURL, dstPath)
	_, err = task.DownloadFile(srcURL, dstPath)
	task.Output(ps.Output, dstPath)
	return err
}

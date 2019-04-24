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
	srcURL, err := evaluate.ExpressionString(ps.Source, task.Inputs)
	if err != nil {
		log.Printf("Expression failed: %s", err)
		return err
	}
	dstPath, err := evaluate.ExpressionString(ps.Dest, task.Inputs)
	task.Printf("Downloading: %s to %s", srcURL, dstPath)
	_, err = task.DownloadFile(srcURL, dstPath)
	task.Output(ps.Output, dstPath)
	return err
}

/*
func (ps *CopyFileStep) Run(task *Task) error {
	if ps.ArgsCopy != "" {
		dstPath := path.Join(task.Workdir, ps.ArgsCopy.Dest)
		srcPath := task.Inputs[ps.ArgsCopy.Source]
		log.Printf("Copy %s to %s", srcPath, dstPath)
		cpCmd := exec.Command("cp", "-rf", srcPath, dstPath)
		err := cpCmd.Run()
		return err
	}
}
*/

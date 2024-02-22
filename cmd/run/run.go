package run

import (
	"os"
	"path/filepath"

	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
)

func ExecuteFile(playFile string, workDir string, outDir string, inputs map[string]string) error {
	logger.Info("Starting", "playFile", playFile)
	pb := playbook.Playbook{}
	if err := playbook.ParseFile(playFile, &pb); err != nil {
		logger.Error("%s", err)
		return err
	}
	a, _ := filepath.Abs(playFile)
	baseDir := filepath.Dir(a)
	logger.Debug("parsed file", "baseDir", baseDir, "playbook", pb)
	return Execute(pb, baseDir, workDir, outDir, inputs)
}

func Execute(pb playbook.Playbook, baseDir string, workDir string, outDir string, inputs map[string]string) error {

	if outDir == "" {
		outDir = pb.GetDefaultOutDir()
	}

	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		os.MkdirAll(outDir, 0777)
	}

	nInputs, err := pb.PrepConfig(inputs, workDir)
	if err != nil {
		return err
	}
	logger.Debug("Running", "outDir", outDir)

	t := task.NewTask(pb.Name, baseDir, workDir, outDir, nInputs)
	err = pb.Execute(t)
	return err
}

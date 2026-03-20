package run

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
)

func ExecuteFile(playFile string, workDir string, outDir string, inputs map[string]string, debugDir string, debugLimit int) error {
	logger.Info("Starting", "playFile", playFile)
	pb := playbook.Playbook{}
	if err := playbook.ParseFile(playFile, &pb); err != nil {
		logger.Error("%s", err)
		return err
	}
	a, _ := filepath.Abs(playFile)
	baseDir := filepath.Dir(a)
	logger.Debug("parsed file", "baseDir", baseDir, "playbook", pb)
	return Execute(pb, baseDir, workDir, outDir, inputs, debugDir, debugLimit)
}

func Execute(pb playbook.Playbook, baseDir string, workDir string, outDir string, params map[string]string, debugDir string, debugLimit int) error {

	if outDir == "" {
		outDir = pb.GetDefaultOutDir()
	}

	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		os.MkdirAll(outDir, 0777)
	}

	// Setup debug capture directory if enabled
	// Enable if: user explicitly set dir, OR user changed limit from default
	enableDebug := debugDir != "" || (debugLimit != 10)
	if enableDebug {
		if debugDir == "" {
			debugDir = filepath.Join(workDir, "debug-capture")
		} else if !filepath.IsAbs(debugDir) {
			debugDir = filepath.Join(workDir, debugDir)
		}
		if info, err := os.Stat(debugDir); err != nil {
			if os.IsNotExist(err) {
				if mkErr := os.MkdirAll(debugDir, 0777); mkErr != nil {
					logger.Error("Failed to create debug directory", "error", mkErr)
					return mkErr
				}
			} else {
				logger.Error("Failed to access debug directory", "path", debugDir, "error", err)
				return err
			}
		} else if !info.IsDir() {
			logger.Error("Debug path exists but is not a directory", "path", debugDir)
			return fmt.Errorf("debug path %s exists but is not a directory", debugDir)
		}
		logger.Info("Debug capture enabled", "dir", debugDir, "limit", debugLimit)
	}

	nInputs, err := pb.PrepConfig(params, workDir)
	if err != nil {
		return err
	}
	logger.Debug("Running", "outDir", outDir)

	t := task.NewTask(pb.Name, baseDir, workDir, outDir, nInputs)
	err = pb.ExecuteWithCapture(t, debugDir, debugLimit)
	return err
}

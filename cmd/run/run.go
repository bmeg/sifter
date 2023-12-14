package run

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
)

func ExecuteFile(playFile string, workDir string, outDir string, inputs map[string]string) error {
	log.Printf("Starting: %s\n", playFile)
	log.Default().SetPrefix(fmt.Sprintf("%s: ", playFile))
	pb := playbook.Playbook{}
	if err := playbook.ParseFile(playFile, &pb); err != nil {
		log.Printf("%s", err)
		return err
	}
	a, _ := filepath.Abs(playFile)
	baseDir := filepath.Dir(a)
	log.Printf("basedir: %s", baseDir)
	log.Printf("playbook: %s", pb)
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
	log.Printf("Outdir: %s", outDir)

	t := task.NewTask(pb.Name, baseDir, workDir, outDir, nInputs)
	err = pb.Execute(t)
	return err
}

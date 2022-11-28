package run

import (
	"log"
	"os"

	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
)

func Execute(playFile string, workDir string, outDir string, inputs map[string]string) error {
	log.Printf("Starting: %s\n", playFile)
	pb := playbook.Playbook{}
	if err := playbook.ParseFile(playFile, &pb); err != nil {
		log.Printf("%s", err)
		return err
	}

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
	t := task.NewTask(pb.Name, workDir, outDir, nInputs)
	err = pb.Execute(t)
	return err
}

package run

import (
	"fmt"
	"log"
	"os"

	"github.com/bmeg/sifter/playbook"
)

func Execute(playFile string, workDir string, outDir string, inputs map[string]interface{}) error {

	fmt.Printf("Starting: %s\n", playFile)
	pb := playbook.Playbook{}
	if err := playbook.ParseFile(playFile, &pb); err != nil {
		log.Printf("%s", err)
		return err
	}

	if outDir == "" {
		outDir = pb.GetOutdir()
	}

	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		os.MkdirAll(outDir, 0777)
	}

	nInputs := pb.PrepInputs(inputs, workDir)
	m := playbook.Manager{}
	err := pb.Execute(&m, nInputs, workDir, outDir)
	return err
}

package run

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/playbook"
)

func Execute(playFile string, workDir string, outDir string, inputs map[string]interface{}) error {
	fmt.Printf("Starting: %s\n", playFile)
	pb := playbook.Playbook{}
	if err := playbook.ParseFile(playFile, &pb); err != nil {
		log.Printf("%s", err)
		return err
	}
	m := playbook.Manager{}
	err := pb.Execute(&m, inputs, workDir, outDir)
	return err
}

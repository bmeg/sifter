package run

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/playbook"
)

func Execute(playFile string, workDir string, outDir string, inputs map[string]interface{}, man *manager.Manager) error {
	fmt.Printf("Starting: %s\n", playFile)
	pb := playbook.Playbook{}
	if err := playbook.ParseFile(playFile, &pb); err != nil {
		log.Printf("%s", err)
		return err
	}
	err := pb.Execute(man, inputs, workDir, outDir)
	return err
}

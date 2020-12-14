package run

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/manager"
)

func Execute(playFile string, workDir string, inputs map[string]interface{}, man *manager.Manager) error {
	fmt.Printf("Starting: %s\n", playFile)
	pb := manager.Playbook{}
	if err := manager.ParseFile(playFile, &pb); err != nil {
		log.Printf("%s", err)
		return err
	}
	err := pb.Execute(man, inputs, workDir)
	return err
}

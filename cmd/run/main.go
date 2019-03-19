package run

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/manager"

	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run importer",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		man, err := manager.Init()
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}

		playFile := args[0]

		fmt.Printf("Starting: %s\n", playFile)
		pb := manager.Playbook{}
		if err := manager.ParseFile(playFile, &pb); err != nil {
			log.Printf("%s", err)
		}

		for _, step := range pb.Steps {
			if step.MatrixLoad != nil {
				log.Printf("%s\n", step.Desc)
				elemStream := step.MatrixLoad.Load()
				for elem := range elemStream {
					log.Printf("%s", elem)
				}
			}
			if step.ManifestLoad != nil {
				log.Printf("Manifest %s\n", step.Desc)
				elemStream := step.ManifestLoad.Load(man.NewTask(map[string]interface{}{}))
				for elem := range elemStream {
					log.Printf("%s", elem)
				}
			}
		}
		return nil
	},
}

func init() {

}

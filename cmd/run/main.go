package run

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/playbook"

	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run importer",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		pb := playbook.Playbook{}
		man, err := manager.Init(args[1:])
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}

		playFile := args[0]

		fmt.Printf("Starting: %s\n", playFile)

		if err := playbook.ParseFile(playFile, &pb); err != nil {
			log.Printf("%s", err)
		}

		for _, prep := range pb.Prep {
			prep.Run(man)
		}

		//fmt.Printf("%s", pb)

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
				elemStream := step.ManifestLoad.Load(man)
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

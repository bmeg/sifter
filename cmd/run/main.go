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

		pb.Execute(man)
		return nil
	},
}

func init() {

}

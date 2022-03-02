package manifest

import (
	"fmt"
	//"log"
	//"io/ioutil"

	//"github.com/bmeg/sifter/steps"
	"github.com/bmeg/sifter/manifest"

	"github.com/spf13/cobra"
)

// CheckCmd is the declaration of the command line
var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check files against manifest",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		man, err := manifest.Load(args[0])
		if err != nil {
			return err
		}
		for _, ent := range man.Entries {
			state := "MISSING"
			fileMD5 := ""
			if ent.Exists() {
				state = "OK"
				if m, err := ent.CalcMD5(); err == nil {
					fileMD5 = m
				} else {
					fileMD5 = "ERROR"
				}
			}
			manMD5 := ent.MD5
			if ent.MD5 == "" {
				manMD5 = "NoMD5"
			}
			if state == "OK" {
				if manMD5 != fileMD5 {
					state = "MD5-Change"
				}
			}
			fmt.Printf("%s\t%s\t%s\t%s\n", ent.Path, state, fileMD5, manMD5)
		}
		return nil
	},
}

var SumCmd = &cobra.Command{
	Use:   "sum",
	Short: "Compute and update checksums for manifest",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		man, err := manifest.Load(args[0])
		if err != nil {
			return err
		}
		changed := false
		for i := range man.Entries {
			if man.Entries[i].Exists() {
				if m, err := man.Entries[i].CalcMD5(); err == nil {
					if m != man.Entries[i].MD5 {
						fmt.Printf("Updating %s\n", man.Entries[i].Path)
						man.Entries[i].MD5 = m
						changed = true
					}
				}
			}
		}
		if changed {
			fmt.Printf("Saving results\n")
		}
		return nil
	},
}

var Cmd = &cobra.Command{
	Use:           "manifest",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	Cmd.AddCommand(CheckCmd)
	Cmd.AddCommand(SumCmd)
}

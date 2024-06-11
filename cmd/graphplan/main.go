package graphplan

import (
	"path/filepath"

	"github.com/bmeg/sifter/graphplan"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/playbook"
	"github.com/spf13/cobra"
)

var outScriptDir = ""
var outDataDir = "./"
var objectExclude = []string{}
var verbose bool = false

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph-plan",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if verbose {
			logger.Init(true, false)
		}

		scriptPath, _ := filepath.Abs(args[0])

		/*
			if outScriptDir != "" {
				baseDir, _ = filepath.Abs(outScriptDir)
			} else if len(args) > 1 {
				return fmt.Errorf("for multiple input directories, based dir must be defined")
			}

			_ = baseDir
		*/
		outScriptDir, _ = filepath.Abs(outScriptDir)
		outDataDir, _ = filepath.Abs(outDataDir)

		outDataDir, _ = filepath.Rel(outScriptDir, outDataDir)

		pb := playbook.Playbook{}

		if sifterErr := playbook.ParseFile(scriptPath, &pb); sifterErr == nil {
			if len(pb.Pipelines) > 0 || len(pb.Inputs) > 0 {
				err := graphplan.NewGraphBuild(
					&pb, outScriptDir, outDataDir, objectExclude,
				)
				if err != nil {
					logger.Error("Parse Error", "error", err)
				}
			}
		}

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVarP(&verbose, "verbose", "v", verbose, "Verbose logging")
	flags.StringVarP(&outScriptDir, "dir", "C", outScriptDir, "Change Directory for script base")
	flags.StringVarP(&outDataDir, "out", "o", outDataDir, "Change output Directory")
	flags.StringArrayVarP(&objectExclude, "exclude", "x", objectExclude, "Object Exclude")
}

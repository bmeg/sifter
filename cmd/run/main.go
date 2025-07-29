package run

import (
	"io"
	"os"

	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/playbook"

	"github.com/spf13/cobra"
)

var outDir string = ""
var inputFile string = ""
var verbose bool = false
var cmdInputs map[string]string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run <script>",
	Short: "Run sifter script",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if verbose {
			logger.Init(true, false)
		}

		inputs := map[string]string{}
		if inputFile != "" {
			if err := playbook.ParseStringFile(inputFile, &inputs); err != nil {
				logger.Error("%s", err)
				return err
			}
		}
		for k, v := range cmdInputs {
			inputs[k] = v
		}
		for _, playFile := range args {
			if playFile == "-" {
				yaml, err := io.ReadAll(os.Stdin)
				if err != nil {
					logger.Error("%s", err)
					return err
				}
				pb := playbook.Playbook{}
				playbook.ParseBytes(yaml, "./playbook.yaml", &pb)
				if err := Execute(pb, "./", "./", outDir, inputs); err != nil {
					return err
				}
			} else {
				if err := ExecuteFile(playFile, "./", outDir, inputs); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVarP(&verbose, "verbose", "v", verbose, "Verbose logging")
	flags.StringToStringVarP(&cmdInputs, "config", "c", cmdInputs, "Config variable")
	flags.StringVarP(&inputFile, "configFile", "f", inputFile, "Config file")
}

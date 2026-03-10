package run

import (
	"io"
	"os"

	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/playbook"

	"github.com/spf13/cobra"
)

var outDir string = ""
var paramsFile string = ""
var verbose bool = false
var cmdParams map[string]string
var debugOutputDir string = ""
var debugSampleLimit int = 10

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run <script>",
	Short: "Run sifter script",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if verbose {
			logger.Init(true, false)
		}

		params := map[string]string{}
		if paramsFile != "" {
			if err := playbook.ParseStringFile(paramsFile, &params); err != nil {
				logger.Error("%s", err)
				return err
			}
		}
		for k, v := range cmdParams {
			params[k] = v
			logger.Info("Input Params", k, v)
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
				if err := Execute(pb, "./", "./", outDir, params, debugOutputDir, debugSampleLimit); err != nil {
					return err
				}
			} else {
				if err := ExecuteFile(playFile, "./", outDir, params, debugOutputDir, debugSampleLimit); err != nil {
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
	flags.StringToStringVarP(&cmdParams, "param", "p", cmdParams, "Parameter variable")
	flags.StringVarP(&paramsFile, "params-file", "f", paramsFile, "Parameter file")
	flags.StringVarP(&debugOutputDir, "debug-output-dir", "d", "", "Directory for debug capture files (default: ./debug-capture)")
	flags.IntVarP(&debugSampleLimit, "debug-sample-limit", "l", 10, "Max records to capture per step (0 = unlimited)")
}

package scan

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/playbook"
	"github.com/spf13/cobra"
)

var jsonOut = false

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "scan <dir>",
	Short: "Scan for outputs",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir := args[0]

		ScanSifter(baseDir, func(pb *playbook.Playbook) {

			for pname, p := range pb.Pipelines {
				emitName := ""
				for _, s := range p {
					if s.Emit != nil {
						emitName = s.Emit.Name
					}
				}
				if emitName != "" {
					for _, s := range p {
						if s.ObjectValidate != nil {
							outdir := pb.GetDefaultOutDir()
							outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)
							outpath := filepath.Join(outdir, outname)
							//outpath, _ = filepath.Rel(baseDir, outpath)
							fmt.Printf("%s\t%s\n", s.ObjectValidate.Title, outpath)
						}
					}
				}
			}

		})

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVarP(&jsonOut, "json", "j", jsonOut, "Output JSON")
}

func ScanSifter(baseDir string, userFunc func(*playbook.Playbook)) {
	filepath.Walk(baseDir,
		func(path string, info fs.FileInfo, err error) error {
			if strings.HasSuffix(path, ".yaml") {
				pb := playbook.Playbook{}
				if parseErr := playbook.ParseFile(path, &pb); parseErr == nil {
					userFunc(&pb)
				}
			}
			return nil
		})
}

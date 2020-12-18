package template

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/extractors"
	"github.com/bmeg/sifter/manager"
)

var workDir string = "./"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "template",
	Short: "Run templated job",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		outDir := "./"
		driver := fmt.Sprintf("dir://%s", outDir)

		//TODO: This needs to be configurable
		dsConfig := datastore.Config{URL: "mongodb://localhost:27017", Database: "sifter", Collection: "cache"}

		man, err := manager.Init(manager.Config{Driver: driver, WorkDir: workDir, DataStore: &dsConfig})
		if err != nil {
			return err
		}

		template := args[0]
		steps := strings.Split(template, ":")
		inputs := map[string]interface{}{}
		for _, k := range args[1:] {
			tmp := strings.Split(k, "=")
			inputs[tmp[0]] = tmp[1]
		}

		if ext, ok := ExtractTemplates[steps[0]]; ok {
			if trans, ok := TransformTemplates[steps[1]]; ok {
				if dec, ok := ExtractorDecorate[steps[0]]; ok {
					ext = dec(ext, trans)

					dir, err := ioutil.TempDir(workDir, "sifterwork_")
					if err != nil {
						log.Fatal(err)
						return err
					}
					pb := manager.Playbook{
						Name: template,
						Inputs: map[string]manager.Input{
							"input": {Type: "file"},
						},
						Steps: []extractors.Extractor{
							ext,
						},
					}
					err = pb.Execute(man, inputs, dir)
					return err
				}
			}
		}
		return nil
	},
}

func init() {
	//flags := Cmd.Flags()
	//flags.StringVar(&template, "template", template, "Workdir")
}

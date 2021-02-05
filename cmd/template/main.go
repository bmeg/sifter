package template

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/extractors"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/manager"
)

var extractMethod string = ""
var transformMethod string = ""
var loadMethod string = "dir://."

var extractOpts = map[string]string{}
var transformOpts = map[string]string{}
var loadOpts = map[string]string{}

var workDir string = "./"
var cache string = ""

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "template",
	Short: "Run templated job",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {

		//TODO: This needs to be configurable
		var dsConfig *datastore.Config
		if cache != "" {
			dsConfig = &datastore.Config{URL: cache, Database: "sifter", Collection: "cache"}
		}

		var ld loader.Loader
		if lm, ok := LoadTemplates[loadMethod]; ok {
			var err error
			ld, err = lm(loadOpts)
			if err != nil {
				return err
			}
		}
		defer ld.Close()

		man, err := manager.Init(manager.Config{Loader: ld, WorkDir: workDir, DataStore: dsConfig})
		if err != nil {
			return err
		}

		inputs := map[string]interface{}{}
		for k, v := range extractOpts {
			inputs[k] = v
		}
		for k, v := range transformOpts {
			inputs[k] = v
		}

		if ext, ok := ExtractTemplates[extractMethod]; ok {
			if trans, ok := TransformTemplates[transformMethod]; ok {
				if dec, ok := ExtractorDecorate[extractMethod]; ok {
					ext = dec(ext, trans)

					dir, err := ioutil.TempDir(workDir, "sifterwork_")
					if err != nil {
						log.Fatal(err)
						return err
					}
					pb := manager.Playbook{
						Name: fmt.Sprintf("%s:%s:%s", extractMethod, transformMethod, loadMethod),
						Inputs: map[string]manager.Input{
							"input": {Type: "File"},
						},
						Steps: []extractors.Extractor{
							ext,
						},
					}
					err = pb.Execute(man, inputs, dir)
					return err
				}
			} else {
				return fmt.Errorf("Transformer %s not found", transformMethod)
			}
		} else {
			return fmt.Errorf("Extractor %s not found", extractMethod)
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&extractMethod, "extract", "E", extractMethod, "Name of extractor engine to use")
	flags.StringVarP(&transformMethod, "transform", "T", transformMethod, "Name of transform template to use")
	flags.StringVarP(&loadMethod, "load", "L", loadMethod, "Name of load engine to use")

	flags.StringToStringVarP(&extractOpts, "extract-opts", "e", extractOpts, "Options for extractor engine")
	flags.StringToStringVarP(&transformOpts, "transform-opts", "t", transformOpts, "Options for transform template")
	flags.StringToStringVarP(&loadOpts, "load-opts", "l", loadOpts, "Options for load engine")
}

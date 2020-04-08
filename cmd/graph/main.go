package graph

import (
	"log"
	"fmt"
	"strings"
	"path/filepath"
	"github.com/spf13/cobra"
	"encoding/json"

  "github.com/bmeg/sifter/graph"
	"github.com/bmeg/sifter/schema"
	"github.com/bmeg/golib"

)

var outDir  string = "./out-graph"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph",
	Short: "Build graph from object files",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		schemaDir := args[0]
		inDir := args[1]


		schemas, err := schema.Load(schemaDir)
		if err != nil {
      return err
    }

		driver := fmt.Sprintf("dir://%s", outDir)

		builder,err := graph.NewBuilder(driver, schemas)
    if err != nil {
      return err
    }

		paths, _ := filepath.Glob(filepath.Join(inDir, "*.json.gz"))
		for _, path := range paths {
			n := filepath.Base(path)
			log.Printf("%s", n)

			reader, err := golib.ReadGzipLines(path)
			if err == nil {
				objChan := make(chan map[string]interface{}, 100)
				go func() {
					defer close(objChan)
					for line := range reader {
						o := map[string]interface{}{}
						if len(line) > 0 {
							json.Unmarshal(line, &o)
							objChan <- o
						}
					}
				}()
				tmp := strings.Split(n, ".")
				prefix := tmp[0]
				class := tmp[1]
				builder.Process( prefix, class, objChan )
			}
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&outDir, "o", outDir, "Output Dir")
}

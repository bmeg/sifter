package graph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/graphbuild"
	"github.com/bmeg/sifter/schema"
)

var outDir string = "./out-graph"
var workDir string = "./"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph-build [mapping] [inputDir]",
	Short: "Build graph from object files",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		mappingPath := args[0]
		inDir := args[1]

		tmpDir, err := ioutil.TempDir(workDir, "siftergraph_")
		if err != nil {
			return err
		}

		mapping, err := graphbuild.LoadMapping(mappingPath)
		if err != nil {
			return err
		}
		log.Printf("Loaded Mapping: %s", mappingPath)

		schemas, err := schema.Load(mapping.Schema)
		if err != nil {
			return err
		}
		log.Printf("Loaded Schema: %s", mapping.Schema)

		emitter := NewDomainEmitter(outDir, mapping.GetVertexPrefixes(), mapping.GetEdgeEndPrefixes())

		fmt.Printf("%s\n", mapping.GetVertexPrefixes())
		fmt.Printf("%s\n", mapping.GetEdgeEndPrefixes())

		paths, _ := filepath.Glob(filepath.Join(inDir, "*.json.gz"))
		for _, path := range paths {
			if mapping.HasRule( path ) {
				log.Printf("Processing: %s", path)
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
					mapping.Process(path, objChan, schemas, emitter)
				}
			}
		}
		os.RemoveAll(tmpDir)
		emitter.Close()
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&outDir, "out", "o", outDir, "Output Dir")
}

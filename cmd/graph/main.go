package graph

import (
	"encoding/json"
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


func RunGraphBuild(mappingPath string, inputDir string, workdir string, outputDir string) error {
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

	emitter := NewDomainEmitter(outputDir, mapping.GetVertexPrefixes(), mapping.GetEdgeEndPrefixes())

	log.Printf("Vertex Prefixes: %s\n", mapping.GetVertexPrefixes())
	log.Printf("EdgeEdge Prefixes: %s\n", mapping.GetEdgeEndPrefixes())

	paths, _ := filepath.Glob(filepath.Join(inputDir, "*.json.gz"))
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
	emitter.Close()
	return nil
}


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

		err = RunGraphBuild(mappingPath, inDir, tmpDir, outDir)
		os.RemoveAll(tmpDir)
		return err
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&outDir, "out", "o", outDir, "Output Dir")
}

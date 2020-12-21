package graph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bmeg/golib"
	"github.com/bmeg/sifter/graphbuild"
	"github.com/bmeg/sifter/schema"
)

var outDir string = "./out-graph"
var mappingFile string
var workDir string = "./"

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

		tmpDir, err := ioutil.TempDir(workDir, "siftergraph_")
		if err != nil {
			return err
		}

		driver := fmt.Sprintf("dir://%s", outDir)
		builder, err := graphbuild.NewBuilder(driver, schemas, tmpDir)
		if err != nil {
			return err
		}

		if mappingFile != "" {
			m, err := graphbuild.LoadMapping(mappingFile, inDir)
			if err != nil {
				return err
			}
			log.Printf("Loaded Mapping: %s", mappingFile)
			builder.AddMapping(m)
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
				builder.Process(prefix, class, objChan)
			}
		}
		builder.Close()

		builder.Report(outDir)
		os.RemoveAll(tmpDir)
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&outDir, "o", outDir, "Output Dir")
	flags.StringVar(&mappingFile, "m", mappingFile, "Mapping File")
}

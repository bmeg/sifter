package graphfix

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/graphedit"
	"github.com/spf13/cobra"
)

func RemapGraph(config *graphedit.Config, src, dst string) error {

	src, _ = filepath.Abs(src)
	dst, _ = filepath.Abs(dst)

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		os.MkdirAll(dst, 0777)
	}

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".Vertex.json.gz") {
			rel, _ := filepath.Rel(src, path)
			dstPath := filepath.Join(dst, rel)
			config.EditVertexFile(path, dstPath)
		} else if strings.HasSuffix(path, ".Edge.json.gz") {
			rel, _ := filepath.Rel(src, path)
			dstPath := filepath.Join(dst, rel)
			config.EditEdgeFile(path, dstPath)
		}
		return nil
	})
	return err
}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph-fix <fix plan> <in dir> <out dir>",
	Short: "Fix Graph by remapping edges",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ge, _ := graphedit.LoadGraphEdit(args[0])
		return RemapGraph(ge, args[1], args[2])
	},
}

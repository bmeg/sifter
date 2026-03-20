package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
)

//go:embed static/*
var staticFS embed.FS

var playbookDir string
var siteDir string
var port string = "8081"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "web <script>",
	Short: "View sifter script in browser",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		playbookDir = args[0]

		var httpFS http.FileSystem

		if siteDir == "" {
			// Serve embedded static files
			staticFiles, err := fs.Sub(staticFS, "static")
			if err != nil {
				log.Fatalf("failed to create sub FS: %v", err)
			}
			httpFS = http.FS(staticFiles)
		} else {
			httpFS = http.Dir(siteDir)
		}
		http.Handle("/", http.FileServer(httpFS))

		// API endpoints
		ph := playbookHandler{playbookDir}
		http.HandleFunc("/api/files", ph.listFilesHandler)
		http.HandleFunc("/api/playbook", ph.getPlaybookHandler)

		fmt.Printf("Server listening on http://localhost:%s\n", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))

		return nil
	},
}

type playbookHandler struct {
	baseDir string
}

type fileTreeNode struct {
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	IsDir    bool           `json:"isDir"`
	Children []fileTreeNode `json:"children,omitempty"`
}

func buildFileTree(absDir string, relativeDir string) ([]fileTreeNode, error) {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	nodes := make([]fileTreeNode, 0, len(entries))
	for _, entry := range entries {
		entryRelPath := filepath.Join(relativeDir, entry.Name())
		node := fileTreeNode{
			Name:  entry.Name(),
			Path:  filepath.ToSlash(entryRelPath),
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			children, err := buildFileTree(filepath.Join(absDir, entry.Name()), entryRelPath)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (ph *playbookHandler) listFilesHandler(w http.ResponseWriter, r *http.Request) {
	basePath, err := filepath.Abs(ph.baseDir)
	if err != nil {
		http.Error(w, "failed to resolve playbook directory", http.StatusInternalServerError)
		return
	}

	entries, err := buildFileTree(basePath, "")

	if err != nil {
		http.Error(w, "failed to read playbook directory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (ph *playbookHandler) getPlaybookHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}

	// Allow nested relative paths while preventing traversal and absolute paths.
	cleanName := filepath.Clean(name)
	if cleanName == "." || strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
		http.Error(w, "invalid playbook name", http.StatusBadRequest)
		return
	}

	baseAbsPath, err := filepath.Abs(ph.baseDir)
	if err != nil {
		http.Error(w, "failed to resolve playbook directory", http.StatusInternalServerError)
		return
	}

	path := filepath.Join(ph.baseDir, cleanName)
	targetAbsPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "invalid playbook path", http.StatusBadRequest)
		return
	}

	basePrefix := baseAbsPath + string(os.PathSeparator)
	if targetAbsPath != baseAbsPath && !strings.HasPrefix(targetAbsPath, basePrefix) {
		http.Error(w, "invalid playbook name", http.StatusBadRequest)
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "playbook not found", http.StatusNotFound)
		return
	}
	format := r.URL.Query().Get("format")
	if format == "json" {
		jsonBytes, err := yaml.YAMLToJSON(content)
		if err != nil {
			http.Error(w, "failed to convert yaml to json", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&siteDir, "site-dir", "s", siteDir, "Serve Custom site dir")
	flags.StringVarP(&port, "port", "p", port, "Port to listen on")
}

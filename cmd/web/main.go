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
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
)

//go:embed static/*
var staticFS embed.FS

var playbookDir string
var siteDir string

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
		http.HandleFunc("/api/playbooks", ph.listPlaybooksHandler)
		http.HandleFunc("/api/playbook", ph.getPlaybookHandler)

		port := "8081"
		fmt.Printf("Server listening on http://localhost:%s\n", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))

		return nil
	},
}

type playbookHandler struct {
	baseDir string
}

func (ph *playbookHandler) listPlaybooksHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(ph.baseDir)
	if err != nil {
		http.Error(w, "failed to read playbook directory", http.StatusInternalServerError)
		return
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		// Only include yaml files
		if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
			names = append(names, e.Name())
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

func (ph *playbookHandler) getPlaybookHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}
	// Prevent directory traversal
	if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") {
		http.Error(w, "invalid playbook name", http.StatusBadRequest)
		return
	}
	path := filepath.Join(ph.baseDir, name)
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
}

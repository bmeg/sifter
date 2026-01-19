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

	"github.com/spf13/cobra"
)

//go:embed static/*
var staticFS embed.FS

var playbookDir string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "web <script>",
	Short: "View sifter script in browser",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		playbookDir = args[0]
		// Serve embedded static files
		staticFiles, err := fs.Sub(staticFS, "static")
		if err != nil {
			log.Fatalf("failed to create sub FS: %v", err)
		}
		http.Handle("/", http.FileServer(http.FS(staticFiles)))

		// API endpoints
		http.HandleFunc("/api/playbooks", listPlaybooksHandler)
		http.HandleFunc("/api/playbook", getPlaybookHandler)

		port := "8081"
		fmt.Printf("Server listening on http://localhost:%s\n", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))

		return nil
	},
}

func listPlaybooksHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(playbookDir)
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

func getPlaybookHandler(w http.ResponseWriter, r *http.Request) {
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
	path := filepath.Join("examples", name)
	content, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "playbook not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

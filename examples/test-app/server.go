package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed index.html style.css app.js
var staticFS embed.FS

func main() {
	port := os.Getenv("TEST_APP_PORT")
	if port == "" {
		port = "3000"
	}

	// Try serving from local directory first (for development editing), fallback to embedded
	dir, err := os.Getwd()
	if err == nil {
		testAppDir := filepath.Join(dir, "examples", "test-app")
		if _, err := os.Stat(filepath.Join(testAppDir, "index.html")); err == nil {
			fs := http.FileServer(http.Dir(testAppDir))
			fmt.Printf("🚀 SaaSKit Test App running at http://localhost:%s (serving %s)\n", port, testAppDir)
			log.Fatal(http.ListenAndServe(":"+port, fs))
			return
		}
	}

	fs := http.FileServer(http.FS(staticFS))
	fmt.Printf("🚀 SaaSKit Test App running at http://localhost:%s (embedded assets)\n", port)
	log.Fatal(http.ListenAndServe(":"+port, fs))
}

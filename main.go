package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed favicon.ico
var favicon []byte

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	apiKey := os.Getenv("UPLOAD_API_KEY")
	if apiKey == "" {
		log.Fatal("UPLOAD_API_KEY environment variable is required")
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "/var/www/uploads"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://u.fftp.io"
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("failed to create upload directory: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><head><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="icon" href="/favicon.ico"><title>u</title><style>body{margin:2em;font-family:monospace}@media(prefers-color-scheme:dark){html{background:#111;color:#eee}}</style></head><body><pre>u - file upload service</pre></body></html>`)
	})

	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=604800")
		w.Write(favicon)
	})

	mux.HandleFunc("POST /upload", func(w http.ResponseWriter, r *http.Request) {
		// Auth check
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse multipart form (32 MB max)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file field", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Determine filename
		preserveFilename := r.FormValue("preserve_filename") == "true"
		var filename string

		if preserveFilename && header.Filename != "" {
			filename = header.Filename
		} else {
			ext := filepath.Ext(header.Filename)
			if ext == "" {
				ext = ".bin"
			}
			filename = time.Now().Format("2006-01-02_15-04-05") + ext
		}

		// Handle collisions
		destPath := filepath.Join(uploadDir, filename)
		if _, err := os.Stat(destPath); err == nil {
			base := strings.TrimSuffix(filename, filepath.Ext(filename))
			ext := filepath.Ext(filename)
			for i := 1; ; i++ {
				candidate := fmt.Sprintf("%s-%d%s", base, i, ext)
				destPath = filepath.Join(uploadDir, candidate)
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					filename = candidate
					break
				}
			}
		}

		// Save file
		dst, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "failed to save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "failed to write file", http.StatusInternalServerError)
			return
		}

		url := strings.TrimRight(baseURL, "/") + "/" + filename
		log.Printf("uploaded: %s -> %s", header.Filename, url)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, url)
	})

	log.Printf("u listening on :%s (uploads -> %s)", port, uploadDir)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

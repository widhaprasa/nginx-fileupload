package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func basicAuth(w http.ResponseWriter, r *http.Request) bool {

	authUsername := os.Getenv("AUTH_USERNAME")
	authPassword := os.Getenv("AUTH_PASSWORD")

	if authUsername != "" {

		// Basic Auth
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return false
		}
		authData := authHeader[6:]
		decoded, err := base64.StdEncoding.DecodeString(authData)
		if err != nil {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return false
		}

		token := strings.SplitN(string(decoded), ":", 2)
		if len(token) != 2 || authUsername != token[0] || authPassword != token[1] {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return false
		}
	}

	return true
}

func reload(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Basic Auth
	ok := basicAuth(w, r)
	if !ok {
		return
	}

	// Run the reload command for Nginx
	cmd := exec.Command("nginx", "-s", "reload")
	err := cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload Nginx: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	fmt.Fprintf(w, "Successfully reload nginx")
}

func upload(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Basic Auth
	ok := basicAuth(w, r)
	if !ok {
		return
	}

	// Max file size 1MB
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		http.Error(w, "Error parsing the form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := r.FormValue("name")
	if filename == "" {
		filename = handler.Filename
	}

	directory := r.FormValue("directory")
	if directory == "" {
		directory = "uploads"
	}

	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	extension := filepath.Ext(handler.Filename)
	filename = fmt.Sprintf("%s%s", filename, extension)

	tempFile, err := os.Create(fmt.Sprintf("%s/%s", directory, filename))
	if err != nil {
		http.Error(w, "Error creating the file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	// Build response
	fmt.Fprintf(w, "Successfully upload file: %s\n", filename)
}

func delete(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Basic Auth
	ok := basicAuth(w, r)
	if !ok {
		return
	}

	filename := r.FormValue("name")
	if filename == "" {
		http.Error(w, "Name not found", http.StatusBadRequest)
		return
	}

	directory := r.FormValue("directory")
	if directory == "" {
		directory = "uploads"
	}
	path := fmt.Sprintf("%s/%s", directory, filename)

	err := os.Remove(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete file: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	fmt.Fprintf(w, "Successfully delete file: %s\n", path)
}

func setupRoutes() {
	http.HandleFunc("/reload", reload)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/delete", delete)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

func main() {
	setupRoutes()
}

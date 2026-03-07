package reciever

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type WriteRequest struct {
	Contents string `json:"contents"`
	Path     string `json:"path"`
}

type EditRequest struct {
	OldContents string `json:"oldContents"`
	NewContents string `json:"newContents"`
	Path        string `json:"path"`
}

func resolvePath(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", errors.New("empty path")
	}

	raw := strings.TrimSpace(input)
	if raw == "~" || strings.HasPrefix(raw, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to resolve home directory: %w", err)
		}
		if raw == "~" {
			raw = homeDir
		} else {
			raw = filepath.Join(homeDir, strings.TrimPrefix(raw, "~/"))
		}
	}

	cleaned := filepath.Clean(raw)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	return abs, nil
}

func HandleToolsRead(w http.ResponseWriter, r *http.Request) {
	path, err := resolvePath(r.URL.Query().Get("path"))

	if err != nil {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	fmt.Printf("Agent reading file: %s\n", path)

	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"filepath": path,
		"contents": string(data),
	})
}

func HandleToolsWrite(w http.ResponseWriter, r *http.Request) {
	var req WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	rawPath := r.URL.Query().Get("path")
	if strings.TrimSpace(rawPath) == "" {
		rawPath = req.Path
	}
	path, err := resolvePath(rawPath)
	if err != nil {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	contents := req.Contents

	fmt.Printf("Agent writing to file: %s\n", path)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		http.Error(w, fmt.Sprintf("Error creating directory: %v", err), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		http.Error(w, fmt.Sprintf("Error writing file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func HandleToolsEdit(w http.ResponseWriter, r *http.Request) {
	var req EditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	rawPath := r.URL.Query().Get("path")
	if strings.TrimSpace(rawPath) == "" {
		rawPath = req.Path
	}
	path, err := resolvePath(rawPath)
	if err != nil {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	oldContent := req.OldContents
	newContent := req.NewContents
	if oldContent == "" {
		http.Error(w, "Missing old content in request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("AI Agent editing file: %s\n", path)

	// Read existing file
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file for edit: %v", err), http.StatusInternalServerError)
		return
	}

	// Perform the replace
	contentStr := string(data)
	if !strings.Contains(contentStr, oldContent) {
		http.Error(w, "oldContents not found in file", http.StatusBadRequest)
		return
	}

	newContentStr := strings.Replace(contentStr, oldContent, newContent, 1)

	// Write it back
	if err := os.WriteFile(path, []byte(newContentStr), 0644); err != nil {
		http.Error(w, fmt.Sprintf("Error saving edited file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func HandleToolsRestart(w http.ResponseWriter, r *http.Request) {
	toolname := r.URL.Query().Get("toolname")
	if toolname == "" {
		http.Error(w, "Missing 'toolname' query parameter", http.StatusBadRequest)
		return
	}

	fmt.Printf("Agent restarting tool: %s\n", toolname)

	// Execute systemctl restart
	cmd := exec.Command("systemctl", "restart", toolname)
	if err := cmd.Run(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to restart %s: %v", toolname, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func HandleDirEnum(w http.ResponseWriter, r *http.Request) {
	targetPath := r.URL.Query().Get("path")
	if targetPath == "" {
		targetPath = "." // Default to current directory if not provided
	}
	resolvedPath, err := resolvePath(targetPath)
	if err != nil {
		http.Error(w, "Invalid 'path' query parameter", http.StatusBadRequest)
		return
	}

	levelStr := r.URL.Query().Get("level")
	maxDepth, err := strconv.Atoi(levelStr)
	if err != nil || maxDepth < 1 {
		maxDepth = 1 // Default to level 1 if invalid
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Directory listing for: %s (Max Depth: %d)\n", resolvedPath, maxDepth))

	// Walk the directory tree
	filepath.Walk(resolvedPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors like permission denied
		}

		// Calculate current depth
		relPath, _ := filepath.Rel(resolvedPath, path)
		depth := strings.Count(relPath, string(os.PathSeparator))
		if relPath == "." {
			depth = 0
		}

		// Stop going deeper if we hit the level limit
		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Formatting the tree output
		indent := strings.Repeat("  ", depth)
		if info.IsDir() {
			fmt.Fprintf(&sb, "%s📁 %s/\n", indent, info.Name())
		} else {
			fmt.Fprintf(&sb, "%s📄 %s\n", indent, info.Name())
		}
		return nil
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"contents": sb.String(),
	})
}

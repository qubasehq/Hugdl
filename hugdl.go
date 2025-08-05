package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ModelInfo represents a model file from HuggingFace
type ModelInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size,omitempty"`
	Path string `json:"path"`
}

func main() {
	// Command line flags
	var (
		modelName = flag.String("model", "Qwen/Qwen2.5-Coder-0.5B", "Model name (e.g., Qwen/Qwen2.5-Coder-0.5B)")
		outputDir = flag.String("output", "C:\\Users\\user\\hf\\models", "Output directory for downloaded files")
		help      = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	// Show help if requested
	if *help {
		fmt.Println("🚀 hugdl - Fast HuggingFace Model Downloader")
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println("Usage: hugdl [options]")
		fmt.Println("")
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  hugdl -model Qwen/Qwen2.5-Coder-0.5B")
		fmt.Println("  hugdl -model microsoft/DialoGPT-medium")
		fmt.Println("  hugdl -model meta-llama/Llama-2-7b-chat-hf -output D:\\models")
		return
	}

	fmt.Println("🚀 hugdl - Fast HuggingFace Model Downloader")
	fmt.Println(strings.Repeat("=", 50))

	// Configuration
	baseURL := "https://huggingface.co"
	apiURL := "https://huggingface.co/api"

	// Create model directory name
	modelDirName := strings.ReplaceAll(*modelName, "/", "_")
	modelDir := filepath.Join(*outputDir, modelDirName)

	fmt.Printf("📦 Model: %s\n", *modelName)
	fmt.Printf("📁 Output: %s\n", modelDir)
	fmt.Println(strings.Repeat("=", 50))

	// Step 1: Get model file list
	fmt.Println("🔍 Checking available files...")
	files, err := getModelFiles(apiURL, *modelName)
	if err != nil {
		fmt.Printf("❌ Error getting model files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Found %d files\n", len(files))

	// Step 2: Create output directory
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		fmt.Printf("❌ Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Download all files
	fmt.Println("\n📥 Starting downloads...")
	fmt.Println(strings.Repeat("-", 50))

	successCount := 0
	for i, file := range files {
		fmt.Printf("[%d/%d] Downloading %s...\n", i+1, len(files), file.Path)
		
		if err := downloadFile(baseURL, *modelName, modelDir, file); err != nil {
			fmt.Printf("❌ Failed to download %s: %v\n", file.Path, err)
		} else {
			fmt.Printf("✅ Downloaded %s\n", file.Path)
			successCount++
		}
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("🎉 Download complete! %d/%d files downloaded successfully\n", successCount, len(files))
	fmt.Printf("📁 Files saved to: %s\n", modelDir)
}

// getModelFiles fetches the list of files from HuggingFace API
func getModelFiles(apiURL, modelName string) ([]ModelInfo, error) {
	url := fmt.Sprintf("%s/models/%s/tree/main", apiURL, modelName)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch model info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var apiResponse []struct {
		Type string `json:"type"`
		Path string `json:"path"`
		Size int64  `json:"size,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	var files []ModelInfo
	for _, item := range apiResponse {
		if item.Type == "file" {
			files = append(files, ModelInfo{
				Name: filepath.Base(item.Path),
				Type: item.Type,
				Size: item.Size,
				Path: item.Path,
			})
		}
	}

	return files, nil
}

// downloadFile downloads a single file with simple progress
func downloadFile(baseURL, modelName, modelDir string, file ModelInfo) error {
	// Create download URL
	downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s", baseURL, modelName, file.Path)
	
	// Create output file path
	outputPath := filepath.Join(modelDir, file.Name)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to mimic browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")

	// Make request with timeout
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create output file
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Download with simple progress
	fmt.Printf("   📥 Downloading %s (%d bytes)...\n", file.Name, file.Size)
	
	// Copy with progress
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("   ✅ Downloaded %s (%d bytes)\n", file.Name, written)
	return nil
} 
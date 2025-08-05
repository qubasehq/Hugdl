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

	"github.com/schollz/progressbar/v3"
)

// ModelInfo represents a model from HuggingFace
type ModelInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size,omitempty"`
	Path     string `json:"path"`
}

// DownloadConfig holds download configuration
type DownloadConfig struct {
	ModelName    string
	BaseURL      string
	APIURL       string
	OutputDir    string
	ModelDir     string
}

func main() {
	// Command line flags
	var (
		modelName = flag.String("model", "Qwen/Qwen2.5-Coder-0.5B", "Model name (e.g., Qwen/Qwen2.5-Coder-0.5B)")
		outputDir = flag.String("output", "C:\\Users\\sarat\\hf\\models", "Output directory for downloaded files")
		help      = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	
	// Show help if requested
	if *help {
		fmt.Println("🚀 Go Model Downloader (Full Version)")
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println("Usage: go run main.go [options]")
		fmt.Println("")
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go -model Qwen/Qwen2.5-Coder-0.5B")
		fmt.Println("  go run main.go -model microsoft/DialoGPT-medium")
		fmt.Println("  go run main.go -model meta-llama/Llama-2-7b-chat-hf -output D:\\models")
		return
	}

	fmt.Println("🚀 Go Model Downloader (Full Version)")
	fmt.Println(strings.Repeat("=", 50))

	// Configuration
	config := DownloadConfig{
		ModelName: *modelName,
		BaseURL:   "https://huggingface.co",
		APIURL:    "https://huggingface.co/api",
		OutputDir: *outputDir,
	}

	// Create model directory name
	modelDirName := strings.ReplaceAll(config.ModelName, "/", "_")
	config.ModelDir = filepath.Join(config.OutputDir, modelDirName)

	fmt.Printf("📦 Model: %s\n", config.ModelName)
	fmt.Printf("📁 Output: %s\n", config.ModelDir)
	fmt.Println(strings.Repeat("=", 50))

	// Step 1: Get model file list
	fmt.Println("🔍 Checking available files...")
	files, err := getModelFiles(config)
	if err != nil {
		fmt.Printf("❌ Error getting model files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Found %d files\n", len(files))

	// Step 2: Create output directory
	if err := os.MkdirAll(config.ModelDir, 0755); err != nil {
		fmt.Printf("❌ Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Download all files
	fmt.Println("\n📥 Starting downloads...")
	fmt.Println(strings.Repeat("-", 50))

	successCount := 0
	for i, file := range files {
		fmt.Printf("[%d/%d] Downloading %s...\n", i+1, len(files), file.Path)
		
		if err := downloadFile(config, file); err != nil {
			fmt.Printf("❌ Failed to download %s: %v\n", file.Path, err)
		} else {
			fmt.Printf("✅ Downloaded %s\n", file.Path)
			successCount++
		}
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("🎉 Download complete! %d/%d files downloaded successfully\n", successCount, len(files))
	fmt.Printf("📁 Files saved to: %s\n", config.ModelDir)
}

// getModelFiles fetches the list of files from HuggingFace API
func getModelFiles(config DownloadConfig) ([]ModelInfo, error) {
	apiURL := fmt.Sprintf("%s/models/%s/tree/main", config.APIURL, config.ModelName)
	
	resp, err := http.Get(apiURL)
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

// downloadFile downloads a single file with progress bar
func downloadFile(config DownloadConfig, file ModelInfo) error {
	// Create download URL
	downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s", config.BaseURL, config.ModelName, file.Path)
	
	// Create output file path
	outputPath := filepath.Join(config.ModelDir, file.Name)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to mimic browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")

	// Make request
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

	// Create progress bar
	var bar *progressbar.ProgressBar
	if file.Size > 0 {
		bar = progressbar.NewOptions64(
			file.Size,
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(50),
			progressbar.OptionSetDescription(fmt.Sprintf("[cyan][1/1][reset] %s", file.Name)),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
	}

	// Download with progress
	if bar != nil {
		_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	} else {
		_, err = io.Copy(out, resp.Body)
	}

	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
} 
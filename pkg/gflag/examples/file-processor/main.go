// SPDX-License-Identifier: CC0-1.0

// Package main demonstrates the gflag package with a simple file processing tool.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nzions/sharedgolibs/pkg/gflag"
)

func main() {
	// Define flags with both long and short names
	var (
		verbose   = gflag.BoolP("verbose", "v", false, "enable verbose output")
		recursive = gflag.BoolP("recursive", "r", false, "process directories recursively")
		output    = gflag.StringP("output", "o", "", "output file (default: stdout)")
		format    = gflag.StringP("format", "f", "text", "output format (text, json, csv)")
		count     = gflag.IntP("max-count", "c", 0, "maximum number of files to process (0 = unlimited)")
		quiet     = gflag.BoolP("quiet", "q", false, "suppress non-error output")
		help      = gflag.BoolP("help", "h", false, "show this help message")
	)

	// Parse command line arguments
	gflag.Parse()

	// Show help if requested
	if *help {
		showHelp()
		return
	}

	// Get remaining arguments (file/directory paths)
	paths := gflag.Args()

	// Validate arguments
	if len(paths) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no input files or directories specified\n")
		fmt.Fprintf(os.Stderr, "Use --help for usage information\n")
		os.Exit(1)
	}

	// Validate format
	validFormats := map[string]bool{"text": true, "json": true, "csv": true}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Valid formats: text, json, csv\n", *format)
		os.Exit(1)
	}

	// Show configuration if verbose
	if *verbose && !*quiet {
		fmt.Printf("Configuration:\n")
		fmt.Printf("  Verbose: %t\n", *verbose)
		fmt.Printf("  Recursive: %t\n", *recursive)
		fmt.Printf("  Output: %s\n", getOutputDescription(*output))
		fmt.Printf("  Format: %s\n", *format)
		fmt.Printf("  Max Count: %s\n", getMaxCountDescription(*count))
		fmt.Printf("  Quiet: %t\n", *quiet)
		fmt.Printf("  Input Paths: %v\n", paths)
		fmt.Println()
	}

	// Process files
	processor := &FileProcessor{
		Verbose:   *verbose,
		Recursive: *recursive,
		Output:    *output,
		Format:    *format,
		MaxCount:  *count,
		Quiet:     *quiet,
	}

	err := processor.Process(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf(`gflag-example - File Processing Tool

USAGE:
    gflag-example [OPTIONS] <file|directory>...

DESCRIPTION:
    A demonstration tool that processes files and directories, showcasing
    the gflag package's support for both short and long flag formats.

OPTIONS:
`)
	gflag.CommandLine.PrintDefaults()
	fmt.Printf(`
EXAMPLES:
    # Process files with verbose output
    gflag-example -v file1.txt file2.txt
    gflag-example --verbose file1.txt file2.txt

    # Process directory recursively with JSON output
    gflag-example -r -f json -o result.json /path/to/dir
    gflag-example --recursive --format=json --output=result.json /path/to/dir

    # Combined short flags (verbose + recursive + quiet)
    gflag-example -vrq --format=csv --max-count=10 /path/to/dir

    # Mixed long and short flags
    gflag-example -v --output=result.txt -r /path/to/dir

FLAG FORMATS SUPPORTED:
    Short flags:      -v, -r, -o filename, -f json
    Long flags:       --verbose, --recursive, --output=filename, --format=json
    Combined short:   -vr, -vrq (for boolean flags)
    Mixed:           -v --output=file --recursive

`)
}

func getOutputDescription(output string) string {
	if output == "" {
		return "stdout"
	}
	return output
}

func getMaxCountDescription(count int) string {
	if count == 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", count)
}

// FileProcessor handles the file processing logic
type FileProcessor struct {
	Verbose   bool
	Recursive bool
	Output    string
	Format    string
	MaxCount  int
	Quiet     bool
	processed int
}

// Process processes the given paths
func (p *FileProcessor) Process(paths []string) error {
	var results []FileInfo

	for _, path := range paths {
		if p.MaxCount > 0 && p.processed >= p.MaxCount {
			if p.Verbose && !p.Quiet {
				fmt.Printf("Reached maximum count (%d), stopping\n", p.MaxCount)
			}
			break
		}

		err := p.processPath(path, &results)
		if err != nil {
			return fmt.Errorf("processing path '%s': %w", path, err)
		}
	}

	return p.outputResults(results)
}

// processPath processes a single path (file or directory)
func (p *FileProcessor) processPath(path string, results *[]FileInfo) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access '%s': %w", path, err)
	}

	if info.IsDir() {
		if p.Recursive {
			return p.processDirectory(path, results)
		} else {
			if p.Verbose && !p.Quiet {
				fmt.Printf("Skipping directory '%s' (use -r for recursive)\n", path)
			}
		}
	} else {
		return p.processFile(path, info, results)
	}

	return nil
}

// processDirectory processes a directory
func (p *FileProcessor) processDirectory(dirPath string, results *[]FileInfo) error {
	if p.Verbose && !p.Quiet {
		fmt.Printf("Processing directory: %s\n", dirPath)
	}

	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if p.MaxCount > 0 && p.processed >= p.MaxCount {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			return p.processFile(path, info, results)
		}

		return nil
	})
}

// processFile processes a single file
func (p *FileProcessor) processFile(filePath string, info os.FileInfo, results *[]FileInfo) error {
	if p.MaxCount > 0 && p.processed >= p.MaxCount {
		return nil
	}

	if p.Verbose && !p.Quiet {
		fmt.Printf("Processing file: %s\n", filePath)
	}

	fileInfo := FileInfo{
		Path:    filePath,
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		IsDir:   info.IsDir(),
	}

	*results = append(*results, fileInfo)
	p.processed++

	return nil
}

// outputResults outputs the results in the specified format
func (p *FileProcessor) outputResults(results []FileInfo) error {
	if !p.Quiet {
		fmt.Printf("Processed %d files\n", len(results))
	}

	if len(results) == 0 {
		return nil
	}

	var output string
	switch p.Format {
	case "json":
		output = p.formatJSON(results)
	case "csv":
		output = p.formatCSV(results)
	default:
		output = p.formatText(results)
	}

	if p.Output == "" {
		fmt.Print(output)
	} else {
		err := os.WriteFile(p.Output, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		if !p.Quiet {
			fmt.Printf("Output written to: %s\n", p.Output)
		}
	}

	return nil
}

// FileInfo represents information about a processed file
type FileInfo struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

func (p *FileProcessor) formatText(results []FileInfo) string {
	var sb strings.Builder
	sb.WriteString("\nFile Processing Results:\n")
	sb.WriteString("========================\n\n")

	for _, file := range results {
		sb.WriteString(fmt.Sprintf("Path: %s\n", file.Path))
		sb.WriteString(fmt.Sprintf("Name: %s\n", file.Name))
		sb.WriteString(fmt.Sprintf("Size: %d bytes\n", file.Size))
		sb.WriteString(fmt.Sprintf("Modified: %s\n", file.ModTime))
		sb.WriteString(fmt.Sprintf("Is Directory: %t\n", file.IsDir))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (p *FileProcessor) formatCSV(results []FileInfo) string {
	var sb strings.Builder
	sb.WriteString("path,name,size,mod_time,is_dir\n")

	for _, file := range results {
		sb.WriteString(fmt.Sprintf("%s,%s,%d,%s,%t\n",
			file.Path, file.Name, file.Size, file.ModTime, file.IsDir))
	}

	return sb.String()
}

func (p *FileProcessor) formatJSON(results []FileInfo) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString(`  "files": [`)
	sb.WriteString("\n")

	for i, file := range results {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf(`    {
      "path": "%s",
      "name": "%s", 
      "size": %d,
      "mod_time": "%s",
      "is_dir": %t
    }`, file.Path, file.Name, file.Size, file.ModTime, file.IsDir))
	}

	sb.WriteString("\n  ],\n")
	sb.WriteString(fmt.Sprintf(`  "total_files": %d`, len(results)))
	sb.WriteString("\n}\n")

	return sb.String()
}

package helpers

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	Key string `json:"key"`
}

func SetEnvVariables(args string) {
	pairs := strings.Split(args, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			fmt.Println("Invalid key-value pair:", pair)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		err := os.Setenv(key, value)
		if err != nil {
			fmt.Printf("Failed to set environment variable %q: %v\n", key, err)
		}
	}
}

// CopyFileToCurrentDir copies a file from the given fs.FS to the current directory where the binary is executed.
func CopyFileToCurrentDir(srcFS embed.FS, fileName string) (string, error) {
	// Determine the destination file path in the current directory
	destinationPath := filepath.Join(".", fileName)

	// Read the content of the file from the embed.FS
	fileContent, err := srcFS.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded file: %w", err)
	}

	// Create the destination file in the current directory
	dstFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Write the file contents to the destination file
	_, err = dstFile.Write(fileContent)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Get the absolute path of the created file
	absPath, err := filepath.Abs(destinationPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

func SetupBinary(binaryPath string) error {
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return err
	}

	return nil
}

func CheckBinary(binaryPath string) error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(binaryPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Binary execution error: %v\n", err)
		fmt.Printf("Standard Error:\n%s\n", stderr.String())
		return err
	}

	return nil
}

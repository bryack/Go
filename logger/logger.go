package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func getWriter(output string) (io.Writer, error) {
	if len(output) == 0 {
		return nil, fmt.Errorf("output destination cannot be empty")
	}

	outputToLower := strings.ToLower(output)

	if outputToLower == "stdout" {
		return os.Stdout, nil
	}

	if outputToLower == "stderr" {
		return os.Stderr, nil
	}

	dir := filepath.Dir(output)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}

	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file %s: %w", output, err)
	}

	return file, nil
}

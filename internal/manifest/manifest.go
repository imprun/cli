package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imprun/cli/internal/validation"
)

const FileName = "windforce.json"

type Summary struct {
	App string `json:"app"`
}

func Load(dir string) (Summary, error) {
	data, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		return Summary{}, fmt.Errorf("read %s: %w", FileName, err)
	}
	return Parse(data)
}

func Parse(data []byte) (Summary, error) {
	var summary Summary
	if err := json.Unmarshal(data, &summary); err != nil {
		return Summary{}, fmt.Errorf("parse %s: %w", FileName, err)
	}
	if !validation.AppKey(summary.App) {
		return Summary{}, fmt.Errorf("invalid app key %q in %s", summary.App, FileName)
	}
	return summary, nil
}

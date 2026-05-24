package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ManifestFilename = "terox.json"

type Manifest struct {
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Variables   []Variable `json:"variables,omitempty"`
}

type Variable struct {
	Name    string   `json:"name"`
	Prompt  string   `json:"prompt,omitempty"`
	Default string   `json:"default,omitempty"`
	Type    string   `json:"type,omitempty"`
	Choices []string `json:"choices,omitempty"`
}

func LoadManifest(templateDir string) (*Manifest, error) {
	path := filepath.Join(templateDir, ManifestFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	for i, v := range m.Variables {
		if v.Name == "" {
			return nil, fmt.Errorf("variable at index %d is missing 'name'", i)
		}
	}
	return &m, nil
}

package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const ZeroHash64 = "0000000000000000000000000000000000000000000000000000000000000000"

type Layer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ProjectMeta struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Founder     string `json:"founder"`
	Description string `json:"description"`
}

type StackConfig struct {
	Frontend string `json:"frontend"`
	Runtime  string `json:"runtime"`
	BaaS     string `json:"baas"`
	Cockpit  string `json:"cockpit"`
}

type CockpitConfig struct {
	TelegramBotHandle string `json:"telegram_bot_handle,omitempty"`
	TelegramOwnerID   string `json:"telegram_owner_id,omitempty"`
}

type Manifest struct {
	Version     string        `json:"version"`
	GeneratedAt time.Time     `json:"generated_at"`
	Project     ProjectMeta   `json:"project"`
	Stack       StackConfig   `json:"stack"`
	Cockpit     CockpitConfig `json:"cockpit,omitempty"`
	Layers      []Layer       `json:"layers"`
	SHA256      string        `json:"sha256"`
}

// LoadAndVerify reads a manifest.json from disk and validates its integrity seal.
func LoadAndVerify(filePath string) (*Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading manifest file: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest json: %w", err)
	}

	// Verify self-seal
	expectedHash := m.SHA256
	m.SHA256 = ZeroHash64

	marshaled, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("re-marshaling manifest: %w", err)
	}

	cleanData := strings.TrimRight(string(marshaled), "\r\n")
	sum := sha256.Sum256([]byte(cleanData))
	computedHash := hex.EncodeToString(sum[:])

	if computedHash != expectedHash {
		return nil, fmt.Errorf("cryptographic seal mismatch: expected %s, got %s", expectedHash, computedHash)
	}

	m.SHA256 = expectedHash
	return &m, nil
}

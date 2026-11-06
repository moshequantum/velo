package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ZeroHash64 = "0000000000000000000000000000000000000000000000000000000000000000"

// FileEntry represents an entry in the files[] list of OpenPassport v1.
type FileEntry struct {
	Path       string `json:"path"`
	Layer      string `json:"layer,omitempty"`
	MediaType  string `json:"mediaType,omitempty"`
	SHA256     string `json:"sha256"`
	Visibility string `json:"visibility,omitempty"`
}

// ManifestSelfHash defines the self-hash calculation protocol in OpenPassport v1.
type ManifestSelfHash struct {
	HashMode    string `json:"hashMode"`
	Path        string `json:"path"`
	Placeholder string `json:"placeholder"`
}

// Layer provides backward compatibility for legacy callers.
type Layer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// ProjectMeta provides backward compatibility for legacy callers.
type ProjectMeta struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Founder     string `json:"founder"`
	Description string `json:"description"`
}

// StackConfig provides backward compatibility for legacy callers.
type StackConfig struct {
	Frontend string `json:"frontend"`
	Runtime  string `json:"runtime"`
	BaaS     string `json:"baas"`
	Cockpit  string `json:"cockpit"`
}

// Manifest represents the OpenPassport v1 passport context.
type Manifest struct {
	SchemaVersion      int              `json:"schemaVersion"`
	Kind               string           `json:"kind"`
	Name               string           `json:"name"`
	PackageVersion     string           `json:"packageVersion"`
	Originator         string           `json:"originator,omitempty"`
	TechnicalAuthority string           `json:"technicalAuthority,omitempty"`
	Language           string           `json:"language,omitempty"`
	ManifestSelfHash   ManifestSelfHash `json:"manifestSelfHash"`
	Files              []FileEntry      `json:"files"`

	// VerifiedFiles records the list of all files that passed integrity checks.
	VerifiedFiles []string `json:"-"`

	// Compatibility fields for legacy CLI callers
	Version string      `json:"-"`
	Project ProjectMeta `json:"-"`
	Stack   StackConfig `json:"-"`
	Layers  []Layer     `json:"-"`
	SHA256  string      `json:"-"`
}

// ToLF normalizes text to Unix LF line endings with trimmed boundary whitespace and a trailing newline.
func ToLF(data []byte) []byte {
	str := string(data)
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\r", "\n")
	str = strings.TrimSpace(str) + "\n"
	return []byte(str)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// LoadAndVerify reads a passport directory or manifest.json from disk and validates its integrity seal
// according to the OpenPassport v1 specification (Contract C1).
func LoadAndVerify(targetPath string) (*Manifest, error) {
	fi, err := os.Stat(targetPath)
	if err != nil {
		return nil, fmt.Errorf("reading path: %w", err)
	}

	var manifestPath, baseDir string
	if fi.IsDir() {
		manifestPath = filepath.Join(targetPath, "manifest.json")
		baseDir = targetPath
	} else {
		manifestPath = targetPath
		baseDir = filepath.Dir(targetPath)
	}

	rawBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading manifest file: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(rawBytes, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest json: %w", err)
	}

	// Fail-closed checks on required metadata
	if m.Name == "" {
		return nil, fmt.Errorf("manifest validation failed: missing required 'name' field")
	}
	if m.PackageVersion == "" {
		return nil, fmt.Errorf("manifest validation failed: missing required 'packageVersion' field")
	}
	if len(m.Files) == 0 {
		return nil, fmt.Errorf("manifest validation failed: 'files' list is empty")
	}

	placeholder := m.ManifestSelfHash.Placeholder
	if placeholder == "" {
		placeholder = ZeroHash64
	}

	manifestPathInFiles := m.ManifestSelfHash.Path
	if manifestPathInFiles == "" {
		manifestPathInFiles = "manifest.json"
	}

	// 1. Locate manifest entry in files[]
	var manifestEntry *FileEntry
	for i := range m.Files {
		if m.Files[i].Path == manifestPathInFiles {
			manifestEntry = &m.Files[i]
			break
		}
	}
	if manifestEntry == nil {
		return nil, fmt.Errorf("manifest validation failed: %s entry missing from files list", manifestPathInFiles)
	}

	curHash := strings.ToLower(manifestEntry.SHA256)
	if curHash == "" || len(curHash) != 64 {
		return nil, fmt.Errorf("manifest validation failed: invalid manifest sha256 %q", curHash)
	}

	// 2. Validate manifest self-hash (C1)
	rawStr := string(rawBytes)
	if !strings.Contains(rawStr, curHash) {
		return nil, fmt.Errorf("manifest validation failed: manifest self-hash %s not present in raw file", curHash)
	}
	replaced := strings.Replace(rawStr, curHash, placeholder, 1)
	computedSelfHash := sha256Hex(ToLF([]byte(replaced)))

	if computedSelfHash != curHash {
		return nil, fmt.Errorf("cryptographic seal mismatch for manifest: expected %s, computed %s", curHash, computedSelfHash)
	}

	m.VerifiedFiles = append(m.VerifiedFiles, manifestPathInFiles)

	// 3. Verify all other files listed in files[]
	for _, f := range m.Files {
		if f.Path == manifestPathInFiles {
			continue
		}
		if f.SHA256 == "" {
			return nil, fmt.Errorf("file %q has empty sha256", f.Path)
		}

		filePath := filepath.Join(baseDir, filepath.FromSlash(f.Path))
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("file %q: reading file: %w", f.Path, err)
		}

		computedHash := sha256Hex(ToLF(content))
		expectedHash := strings.ToLower(f.SHA256)
		if computedHash != expectedHash {
			return nil, fmt.Errorf("file %q integrity check failed: expected %s, got %s", f.Path, expectedHash, computedHash)
		}

		m.VerifiedFiles = append(m.VerifiedFiles, f.Path)
	}

	// Populate backward compatibility fields
	m.Version = m.PackageVersion
	m.Project.Name = m.Name
	m.SHA256 = curHash
	for _, f := range m.Files {
		m.Layers = append(m.Layers, Layer{
			ID:     f.Layer,
			Name:   f.Path,
			Path:   f.Path,
			SHA256: f.SHA256,
		})
	}

	return &m, nil
}

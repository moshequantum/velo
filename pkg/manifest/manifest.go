package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const ZeroHash64 = "0000000000000000000000000000000000000000000000000000000000000000"

var (
	ErrPathTraversal = errors.New("path traversal detected: path escapes base directory")
	ErrAbsolutePath  = errors.New("absolute paths are not permitted in passport entries")
)

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

// SafeJoin joins baseDir with an untrusted relative path and ensures that
// the resolved path remains strictly inside baseDir. It rejects ".." traversal,
// absolute paths, and external symlinks.
func SafeJoin(baseDir, untrustedPath string) (string, error) {
	if untrustedPath == "" {
		return "", fmt.Errorf("%w: empty path", ErrPathTraversal)
	}
	if filepath.IsAbs(untrustedPath) || strings.HasPrefix(untrustedPath, "/") || strings.HasPrefix(untrustedPath, "\\") {
		return "", fmt.Errorf("%w: %q", ErrAbsolutePath, untrustedPath)
	}

	cleanRel := filepath.Clean(filepath.FromSlash(untrustedPath))
	if cleanRel == "." || cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q", ErrPathTraversal, untrustedPath)
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}

	target := filepath.Join(absBase, cleanRel)
	relToBase, err := filepath.Rel(absBase, target)
	if err != nil || relToBase == "." || relToBase == ".." || strings.HasPrefix(relToBase, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q", ErrPathTraversal, untrustedPath)
	}

	if fi, err := os.Lstat(target); err == nil && (fi.Mode()&os.ModeSymlink != 0) {
		evaluated, err := filepath.EvalSymlinks(target)
		if err != nil {
			return "", err
		}
		realBase := absBase
		if rb, err := filepath.EvalSymlinks(absBase); err == nil {
			realBase = rb
		}
		symRel, err := filepath.Rel(realBase, evaluated)
		if err != nil || symRel == ".." || strings.HasPrefix(symRel, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("%w (via symlink): %q points outside base directory", ErrPathTraversal, untrustedPath)
		}
	}

	return target, nil
}

// IsValidHex64 reports whether s is a 64-character lowercase hexadecimal string.
func IsValidHex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// IsTextFile reports whether relPath is a text artifact subject to LF normalization.
func IsTextFile(relPath string) bool {
	slash := strings.ReplaceAll(relPath, "\\", "/")
	base := path.Base(slash)
	if base == ".gitattributes" || base == "LICENSE" {
		return true
	}
	ext := strings.ToLower(path.Ext(base))
	return ext == ".md" || ext == ".json" || ext == ".txt" || ext == ".html"
}

// ToLF converts CRLF and lone CR line endings to LF without trimming leading/trailing whitespace (H2).
func ToLF(data []byte) []byte {
	str := string(data)
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\r", "\n")
	return []byte(str)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ComputeManifestSelfHashFromBytes implements the canonical OpenPassport v1.0 self-hash contract (H8):
// It locates the single manifest entry inside the top-level "files" JSON array, replaces only that entry's
// "sha256" string literal matching currentHash with placeholder, normalizes line endings to LF, and returns SHA-256.
func ComputeManifestSelfHashFromBytes(rawContent []byte, manifestEntryPath, currentHash, placeholder string) string {
	normHash := strings.ToLower(strings.TrimSpace(currentHash))
	if !IsValidHex64(normHash) {
		return ""
	}
	if manifestEntryPath == "" {
		manifestEntryPath = "manifest.json"
	}
	if placeholder == "" {
		placeholder = ZeroHash64
	}
	encodedPlaceholder, err := json.Marshal(placeholder)
	if err != nil || len(encodedPlaceholder) < 2 {
		return ""
	}
	placeholderInner := encodedPlaceholder[1 : len(encodedPlaceholder)-1]

	start, end, foundHash, ok := findManifestSha256Span(rawContent, manifestEntryPath)
	if !ok || !strings.EqualFold(foundHash, currentHash) {
		return ""
	}

	replaced := make([]byte, 0, len(rawContent)-(end-start)+len(placeholderInner))
	replaced = append(replaced, rawContent[:start]...)
	replaced = append(replaced, placeholderInner...)
	replaced = append(replaced, rawContent[end:]...)

	return sha256Hex(ToLF(replaced))
}

func skipJSONWhitespace(raw []byte, pos int) int {
	for pos < len(raw) {
		switch raw[pos] {
		case ' ', '\t', '\n', '\r':
			pos++
		default:
			return pos
		}
	}
	return pos
}

func scanJSONString(raw []byte, pos int) (start, end int, val string, ok bool) {
	if pos >= len(raw) || raw[pos] != '"' {
		return 0, 0, "", false
	}
	start = pos
	i := pos + 1
	for i < len(raw) {
		if raw[i] == '\\' {
			i += 2
			continue
		}
		if raw[i] == '"' {
			end = i + 1
			if err := json.Unmarshal(raw[start:end], &val); err != nil {
				return 0, 0, "", false
			}
			return start, end, val, true
		}
		i++
	}
	return 0, 0, "", false
}

func skipJSONValue(raw []byte, pos int) (int, bool) {
	pos = skipJSONWhitespace(raw, pos)
	if pos >= len(raw) {
		return 0, false
	}
	switch raw[pos] {
	case '"':
		_, end, _, ok := scanJSONString(raw, pos)
		return end, ok
	case '{', '[':
		depth := 0
		for pos < len(raw) {
			switch raw[pos] {
			case '"':
				_, end, _, ok := scanJSONString(raw, pos)
				if !ok {
					return 0, false
				}
				pos = end
				continue
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if depth == 0 {
					return pos + 1, true
				}
			}
			pos++
		}
		return 0, false
	default:
		for pos < len(raw) {
			switch raw[pos] {
			case ',', '}', ']', ' ', '\t', '\n', '\r':
				return pos, true
			}
			pos++
		}
		return pos, true
	}
}

func findManifestSha256Span(raw []byte, targetPath string) (valStart, valEnd int, foundHash string, ok bool) {
	pos := skipJSONWhitespace(raw, 0)
	if pos >= len(raw) || raw[pos] != '{' {
		return 0, 0, "", false
	}
	pos++
	for pos < len(raw) {
		pos = skipJSONWhitespace(raw, pos)
		if pos >= len(raw) {
			return 0, 0, "", false
		}
		if raw[pos] == '}' {
			return 0, 0, "", false
		}
		if raw[pos] == ',' {
			pos++
			pos = skipJSONWhitespace(raw, pos)
		}
		_, keyEnd, key, strOk := scanJSONString(raw, pos)
		if !strOk {
			return 0, 0, "", false
		}
		pos = skipJSONWhitespace(raw, keyEnd)
		if pos >= len(raw) || raw[pos] != ':' {
			return 0, 0, "", false
		}
		pos = skipJSONWhitespace(raw, pos+1)
		if key == "files" {
			return scanFilesArrayForManifest(raw, pos, targetPath)
		}
		nextPos, skipOk := skipJSONValue(raw, pos)
		if !skipOk {
			return 0, 0, "", false
		}
		pos = nextPos
	}
	return 0, 0, "", false
}

func scanFilesArrayForManifest(raw []byte, pos int, targetPath string) (valStart, valEnd int, foundHash string, ok bool) {
	pos = skipJSONWhitespace(raw, pos)
	if pos >= len(raw) || raw[pos] != '[' {
		return 0, 0, "", false
	}
	pos++
	matches := 0
	for pos < len(raw) {
		pos = skipJSONWhitespace(raw, pos)
		if pos >= len(raw) {
			return 0, 0, "", false
		}
		if raw[pos] == ']' {
			break
		}
		if raw[pos] == ',' {
			pos++
			pos = skipJSONWhitespace(raw, pos)
		}
		if pos >= len(raw) || raw[pos] != '{' {
			return 0, 0, "", false
		}
		objEnd, isManifest, sStart, sEnd, sVal, objOk := scanFileEntryObject(raw, pos, targetPath)
		if !objOk {
			return 0, 0, "", false
		}
		if isManifest {
			if sStart == 0 && sEnd == 0 {
				return 0, 0, "", false
			}
			matches++
			valStart = sStart
			valEnd = sEnd
			foundHash = sVal
		}
		pos = objEnd
	}
	if matches != 1 {
		return 0, 0, "", false
	}
	return valStart, valEnd, foundHash, true
}

func scanFileEntryObject(raw []byte, pos int, targetPath string) (objEnd int, isManifest bool, shaStart, shaEnd int, shaVal string, ok bool) {
	if pos >= len(raw) || raw[pos] != '{' {
		return 0, false, 0, 0, "", false
	}
	pos++
	for pos < len(raw) {
		pos = skipJSONWhitespace(raw, pos)
		if pos >= len(raw) {
			return 0, false, 0, 0, "", false
		}
		if raw[pos] == '}' {
			return pos + 1, isManifest, shaStart, shaEnd, shaVal, true
		}
		if raw[pos] == ',' {
			pos++
			pos = skipJSONWhitespace(raw, pos)
		}
		_, keyEnd, key, strOk := scanJSONString(raw, pos)
		if !strOk {
			return 0, false, 0, 0, "", false
		}
		pos = skipJSONWhitespace(raw, keyEnd)
		if pos >= len(raw) || raw[pos] != ':' {
			return 0, false, 0, 0, "", false
		}
		pos = skipJSONWhitespace(raw, pos+1)
		switch key {
		case "path":
			_, vEnd, vStr, vOk := scanJSONString(raw, pos)
			if !vOk {
				return 0, false, 0, 0, "", false
			}
			if vStr == targetPath {
				isManifest = true
			}
			pos = vEnd
		case "sha256":
			vStart, vEnd, vStr, vOk := scanJSONString(raw, pos)
			if !vOk {
				return 0, false, 0, 0, "", false
			}
			shaStart = vStart + 1
			shaEnd = vEnd - 1
			shaVal = vStr
			pos = vEnd
		default:
			nextPos, skipOk := skipJSONValue(raw, pos)
			if !skipOk {
				return 0, false, 0, 0, "", false
			}
			pos = nextPos
		}
	}
	return 0, false, 0, 0, "", false
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

	// 1. Locate unique manifest entry in files[]
	var manifestEntry *FileEntry
	manifestCount := 0
	for i := range m.Files {
		if m.Files[i].Path == manifestPathInFiles {
			manifestCount++
			if manifestEntry == nil {
				manifestEntry = &m.Files[i]
			}
		}
	}
	if manifestEntry == nil {
		return nil, fmt.Errorf("manifest validation failed: %s entry missing from files list", manifestPathInFiles)
	}
	if manifestCount != 1 {
		return nil, fmt.Errorf("manifest validation failed: duplicate %s entry in files list", manifestPathInFiles)
	}

	curHash := strings.ToLower(strings.TrimSpace(manifestEntry.SHA256))
	if !IsValidHex64(curHash) {
		return nil, fmt.Errorf("manifest validation failed: invalid manifest sha256 %q", manifestEntry.SHA256)
	}

	// 2. Validate manifest self-hash (C1 + H8)
	computedSelfHash := ComputeManifestSelfHashFromBytes(rawBytes, manifestPathInFiles, manifestEntry.SHA256, placeholder)
	if computedSelfHash == "" || computedSelfHash != curHash {
		return nil, fmt.Errorf("cryptographic seal mismatch for manifest: expected %s, computed %s", curHash, computedSelfHash)
	}

	m.VerifiedFiles = append(m.VerifiedFiles, manifestPathInFiles)

	// 3. Verify all other files listed in files[] concurrently across bounded worker pool
	type entryResult struct {
		path string
		err  error
	}

	jobIndices := make([]int, 0, len(m.Files))
	for i, f := range m.Files {
		if f.Path == manifestPathInFiles {
			continue
		}
		jobIndices = append(jobIndices, i)
	}

	results := make([]entryResult, len(m.Files))
	if len(jobIndices) > 0 {
		jobs := make(chan int, len(jobIndices))
		for _, idx := range jobIndices {
			jobs <- idx
		}
		close(jobs)

		workers := runtime.GOMAXPROCS(0)
		if workers < 1 {
			workers = 1
		}
		if len(jobIndices) < workers {
			workers = len(jobIndices)
		}

		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for idx := range jobs {
					f := m.Files[idx]
					expectedHash := strings.ToLower(strings.TrimSpace(f.SHA256))
					if !IsValidHex64(expectedHash) {
						results[idx] = entryResult{err: fmt.Errorf("file %q has invalid sha256 %q", f.Path, f.SHA256)}
						continue
					}
					filePath, err := SafeJoin(baseDir, f.Path)
					if err != nil {
						results[idx] = entryResult{err: fmt.Errorf("file %q security violation: %w", f.Path, err)}
						continue
					}
					content, err := os.ReadFile(filePath)
					if err != nil {
						results[idx] = entryResult{err: fmt.Errorf("file %q: reading file: %w", f.Path, err)}
						continue
					}
					toHash := content
					if IsTextFile(f.Path) {
						toHash = ToLF(content)
					}
					computedHash := sha256Hex(toHash)
					if computedHash != expectedHash {
						results[idx] = entryResult{err: fmt.Errorf("file %q integrity check failed: expected %s, got %s", f.Path, expectedHash, computedHash)}
						continue
					}
					results[idx] = entryResult{path: f.Path}
				}
			}()
		}
		wg.Wait()
	}

	for _, idx := range jobIndices {
		if results[idx].err != nil {
			return nil, results[idx].err
		}
		m.VerifiedFiles = append(m.VerifiedFiles, results[idx].path)
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

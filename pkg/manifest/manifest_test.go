package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}
}

// Case 1: Golden template verification succeeds against canonical OpenPassport fixture
func TestLoadAndVerify_GoldenVector(t *testing.T) {
	fixtureDir := filepath.Join("testdata", "canonical-template")
	m, err := LoadAndVerify(fixtureDir)
	if err != nil {
		t.Fatalf("expected golden fixture to verify, got error: %v", err)
	}

	const expectedSelfHash = "b6d13a7447ba0abb914f54d338c92440acf5ab151615b5ef2260049c7b2fe5af"
	if m.SHA256 != expectedSelfHash {
		t.Errorf("expected self hash %s, got %s", expectedSelfHash, m.SHA256)
	}

	if m.Name != "OpenPassport Canonical Template" {
		t.Errorf("expected name 'OpenPassport Canonical Template', got %q", m.Name)
	}

	if m.PackageVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", m.PackageVersion)
	}

	if len(m.VerifiedFiles) != len(m.Files) {
		t.Errorf("expected %d verified files, got %d", len(m.Files), len(m.VerifiedFiles))
	}

	// Verify path to manifest.json also works
	manifestPath := filepath.Join(fixtureDir, "manifest.json")
	mDirect, err := LoadAndVerify(manifestPath)
	if err != nil {
		t.Fatalf("expected direct manifest path to verify, got: %v", err)
	}
	if mDirect.SHA256 != expectedSelfHash {
		t.Errorf("expected self hash %s, got %s", expectedSelfHash, mDirect.SHA256)
	}
}

// Case 2: Altered byte in a tracked file (01-IDENTITY.md) must fail with error naming the file
func TestLoadAndVerify_AlteredFileFailsNamingFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join("testdata", "canonical-template")
	copyDir(t, src, tmpDir)

	identityPath := filepath.Join(tmpDir, "01-IDENTITY.md")
	content, err := os.ReadFile(identityPath)
	if err != nil {
		t.Fatalf("failed reading 01-IDENTITY.md: %v", err)
	}

	// Tamper with content
	tampered := append(content, []byte("\n# TAMPERED\n")...)
	if err := os.WriteFile(identityPath, tampered, 0644); err != nil {
		t.Fatalf("failed writing tampered 01-IDENTITY.md: %v", err)
	}

	_, err = LoadAndVerify(tmpDir)
	if err == nil {
		t.Fatal("expected error on tampered file, got nil")
	}

	if !strings.Contains(err.Error(), "01-IDENTITY.md") {
		t.Errorf("expected error message to mention '01-IDENTITY.md', got: %v", err)
	}
}

// Case 3: Altered self-hash in manifest.json must fail verification
func TestLoadAndVerify_AlteredSelfHashFails(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join("testdata", "canonical-template")
	copyDir(t, src, tmpDir)

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed reading manifest.json: %v", err)
	}

	// Tamper with self-hash
	corrupted := bytes.Replace(content, []byte("b6d13a7447ba0abb914f54d338c92440acf5ab151615b5ef2260049c7b2fe5af"), []byte("0000000000000000000000000000000000000000000000000000000000000001"), 1)
	if err := os.WriteFile(manifestPath, corrupted, 0644); err != nil {
		t.Fatalf("failed writing corrupted manifest: %v", err)
	}

	_, err = LoadAndVerify(tmpDir)
	if err == nil {
		t.Fatal("expected error on corrupted self-hash, got nil")
	}

	if !strings.Contains(err.Error(), "cryptographic seal mismatch") && !strings.Contains(err.Error(), "manifest self-hash") {
		t.Errorf("expected cryptographic mismatch error, got: %v", err)
	}
}

// Case 4: CRLF endings on disk must normalize to LF and succeed
func TestLoadAndVerify_CRLFNormalizedSucceeds(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join("testdata", "canonical-template")
	copyDir(t, src, tmpDir)

	// Convert all markdown files in tmpDir to CRLF
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("reading tmpDir: %v", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".md") {
			filePath := filepath.Join(tmpDir, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("reading %s: %v", filePath, err)
			}
			crlfContent := bytes.ReplaceAll(content, []byte("\n"), []byte("\r\n"))
			if err := os.WriteFile(filePath, crlfContent, 0644); err != nil {
				t.Fatalf("writing crlf %s: %v", filePath, err)
			}
		}
	}

	m, err := LoadAndVerify(tmpDir)
	if err != nil {
		t.Fatalf("expected CRLF normalization to succeed, got error: %v", err)
	}

	if len(m.VerifiedFiles) == 0 {
		t.Error("expected verified files to be populated")
	}
}

// Case 5: Missing manifest.json entry in files[] must fail closed
func TestLoadAndVerify_MissingManifestEntryFails(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join("testdata", "canonical-template")
	copyDir(t, src, tmpDir)

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("reading manifest.json: %v", err)
	}

	// Remove manifest.json entry from files array
	if err := os.WriteFile(manifestPath, bytes.Replace(content, []byte(`"path": "manifest.json",`), []byte(`"path": "removed.json",`), 1), 0644); err != nil {
		t.Fatalf("failed replacing manifest path: %v", err)
	}

	_, err = LoadAndVerify(tmpDir)
	if err == nil {
		t.Fatal("expected error when manifest.json entry is missing from files, got nil")
	}

	if !strings.Contains(err.Error(), "manifest.json entry missing") && !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected fail-closed error, got: %v", err)
	}
}

// Case 6: Path traversal ("../secret.txt") in files[] must be rejected by SafeJoin
func TestLoadAndVerify_RejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	baseDir := filepath.Join(root, "passport")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("mkdir baseDir: %v", err)
	}
	copyDir(t, filepath.Join("testdata", "canonical-template"), baseDir)

	secret := []byte("outside-secret\n")
	if err := os.WriteFile(filepath.Join(root, "secret.txt"), secret, 0644); err != nil {
		t.Fatalf("write secret.txt: %v", err)
	}

	manifestPath := filepath.Join(baseDir, "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	m.Files = append(m.Files, FileEntry{
		Path:      "../secret.txt",
		Layer:     "custom",
		MediaType: "text/plain",
		SHA256:    sha256Hex(secret),
	})
	for i := range m.Files {
		if m.Files[i].Path == "manifest.json" {
			m.Files[i].SHA256 = ZeroHash64
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(&m)
	selfHash := sha256Hex(ToLF(buf.Bytes()))
	for i := range m.Files {
		if m.Files[i].Path == "manifest.json" {
			m.Files[i].SHA256 = selfHash
		}
	}
	buf.Reset()
	_ = enc.Encode(&m)
	if err := os.WriteFile(manifestPath, ToLF(buf.Bytes()), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err = LoadAndVerify(baseDir)
	if err == nil || !strings.Contains(err.Error(), "security violation") {
		t.Fatalf("expected SafeJoin security violation error, got: %v", err)
	}
}

// Case 7: Binary PNG containing 0x0d (CR) bytes must hash as raw bytes without corruption
func TestLoadAndVerify_BinaryAssetRawBytes(t *testing.T) {
	tmpDir := t.TempDir()
	copyDir(t, filepath.Join("testdata", "canonical-template"), tmpDir)

	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x0D, 0x00, 0xFF}
	if err := os.WriteFile(filepath.Join(tmpDir, "logo.png"), pngBytes, 0644); err != nil {
		t.Fatalf("write logo.png: %v", err)
	}

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	m.Files = append(m.Files, FileEntry{
		Path:      "logo.png",
		Layer:     "custom",
		MediaType: "image/png",
		SHA256:    sha256Hex(pngBytes),
	})
	for i := range m.Files {
		if m.Files[i].Path == "manifest.json" {
			m.Files[i].SHA256 = ZeroHash64
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(&m)
	selfHash := sha256Hex(ToLF(buf.Bytes()))
	for i := range m.Files {
		if m.Files[i].Path == "manifest.json" {
			m.Files[i].SHA256 = selfHash
		}
	}
	buf.Reset()
	_ = enc.Encode(&m)
	if err := os.WriteFile(manifestPath, ToLF(buf.Bytes()), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	if _, err := LoadAndVerify(tmpDir); err != nil {
		t.Fatalf("expected binary PNG with 0x0D byte to verify cleanly, got: %v", err)
	}
}

// Case 8: H8 strict self-hash targets only the manifest.json entry in files[]
func TestH8_StrictManifestSelfHash(t *testing.T) {
	dupHash := sha256Hex([]byte("duplicate-hash-value"))
	rawTpl := `{
  "schemaVersion": 1,
  "originator": %q,
  "files": [
    {"path": "01-IDENTITY.md", "sha256": %q},
    {"path": "manifest.json", "sha256": %q}
  ]
}
`
	raw := []byte(fmt.Sprintf(rawTpl, dupHash, dupHash, dupHash))
	expectedZeroed := []byte(fmt.Sprintf(rawTpl, dupHash, dupHash, ZeroHash64))
	got := ComputeManifestSelfHashFromBytes(raw, "manifest.json", dupHash, ZeroHash64)
	want := sha256Hex(ToLF(expectedZeroed))
	if got != want {
		t.Fatalf("ComputeManifestSelfHashFromBytes replaced wrong field: got %s, want %s", got, want)
	}
}

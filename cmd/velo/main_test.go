package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/moshequantum/velo/internal/wizard"
)

func captureOutput(f func()) (string, string) {
	origOut := os.Stdout
	origErr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		outC <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	f()

	wOut.Close()
	wErr.Close()

	stdoutStr := <-outC
	stderrStr := <-errC

	os.Stdout = origOut
	os.Stderr = origErr

	return stdoutStr, stderrStr
}

func TestSlugify_Table(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Mi App", "mi-app"},
		{"  Hello World  ", "hello-world"},
		{"Project---Name", "project-name"},
		{"My App! 2026 #Special", "my-app-2026-special"},
		{"simple", "simple"},
		{"Already-Kebab", "already-kebab"},
		{"UPPERCASE_AND_UNDERSCORE", "uppercase_and_underscore"},
	}

	for _, tt := range tests {
		got := wizard.Slugify(tt.input)
		if got != tt.expected {
			t.Errorf("Slugify(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestHandleInit_FlagParsingTable(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCode int
		wantInErr    string
	}{
		{
			name:         "invalid lang enum",
			args:         []string{"--name", "TestApp", "--lang", "fr"},
			expectedCode: 2,
			wantInErr:    "valor inválido para --lang",
		},
		{
			name:         "valid lang es",
			args:         []string{"--name", "TestApp", "--lang", "es", "--dry-run"},
			expectedCode: 0,
		},
		{
			name:         "valid lang en",
			args:         []string{"--name", "TestApp", "--lang", "en", "--dry-run"},
			expectedCode: 0,
		},
		{
			name:         "invalid frontend enum",
			args:         []string{"--name", "TestApp", "--frontend", "angular"},
			expectedCode: 2,
			wantInErr:    "valor inválido para --frontend",
		},
		{
			name:         "invalid runtime enum",
			args:         []string{"--name", "TestApp", "--runtime", "aws-lambda"},
			expectedCode: 2,
			wantInErr:    "valor inválido para --runtime",
		},
		{
			name:         "invalid data enum",
			args:         []string{"--name", "TestApp", "--data", "mongodb"},
			expectedCode: 2,
			wantInErr:    "valor inválido para --data",
		},
		{
			name:         "invalid cockpit enum",
			args:         []string{"--name", "TestApp", "--cockpit", "discord"},
			expectedCode: 2,
			wantInErr:    "valor inválido para --cockpit",
		},
		{
			name:         "valid headless options with data-url",
			args:         []string{"--name", "My App", "--frontend", "astro", "--data", "insforge", "--dry-run"},
			expectedCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var code int
			_, stderrStr := captureOutput(func() {
				code = handleInit(tt.args)
			})

			if code != tt.expectedCode {
				t.Errorf("handleInit(%v) returned code %d, want %d", tt.args, code, tt.expectedCode)
			}
			if tt.wantInErr != "" && !strings.Contains(stderrStr, tt.wantInErr) {
				t.Errorf("handleInit(%v) stderr = %q, want containing %q", tt.args, stderrStr, tt.wantInErr)
			}
		})
	}
}

// DoD: `velo init --name "Mi App" --dry-run` imprime `mi-app/…`
func TestHandleInit_DryRunPrintsSlugPrefix(t *testing.T) {
	args := []string{"--name", "Mi App", "--data", "insforge", "--dry-run"}

	var code int
	stdoutStr, _ := captureOutput(func() {
		code = handleInit(args)
	})

	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}

	expectedPrefix := "mi-app/"
	lines := strings.Split(stdoutStr, "\n")
	hasPrefixLine := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, expectedPrefix) {
			hasPrefixLine = true
			break
		}
	}

	if !hasPrefixLine {
		t.Errorf("expected at least one line starting with %q in dry-run output:\n%s", expectedPrefix, stdoutStr)
	}
}

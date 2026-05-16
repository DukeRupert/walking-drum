package envfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "FOO=from_file\nBAR=also_from_file\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// Make sure these aren't set from the outside.
	t.Setenv("FOO", "")
	os.Unsetenv("FOO")
	t.Setenv("BAR", "")
	os.Unsetenv("BAR")

	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := os.Getenv("FOO"); got != "from_file" {
		t.Errorf("FOO = %q, want %q", got, "from_file")
	}
	if got := os.Getenv("BAR"); got != "also_from_file" {
		t.Errorf("BAR = %q, want %q", got, "also_from_file")
	}
}

func TestLoad_SkipsExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("FOO=from_file\n"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	t.Setenv("FOO", "from_shell")

	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := os.Getenv("FOO"); got != "from_shell" {
		t.Errorf("FOO = %q, want %q (shell value should win)", got, "from_shell")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	err := Load(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParse(t *testing.T) {
	input := `# database config
DATABASE_URL=postgres://user:password@localhost:5432/wd

# server config
PORT=8080
LOG_LEVEL=debug
`
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]string{
		"DATABASE_URL": "postgres://user:password@localhost:5432/wd",
		"PORT":         "8080",
		"LOG_LEVEL":    "debug",
	}

	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %v", len(got), len(want), got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("got[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestParse_ErrorIncludesLineNumber(t *testing.T) {
	input := "FOO=bar\nBROKEN\nBAZ=qux\n"
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("error should mention line 2, got: %v", err)
	}
}

func TestParse_Empty(t *testing.T) {
	got, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantKey string
		wantVal string
		wantOk  bool
		wantErr bool
	}{
		{"simple", "FOO=bar", "FOO", "bar", true, false},
		{"with spaces around equals", "  FOO = bar  ", "FOO", "bar", true, false},
		{"value with equals", "URL=postgres://u:p=w@h/d", "URL", "postgres://u:p=w@h/d", true, false},
		{"comment", "# this is a comment", "", "", false, false},
		{"blank", "   ", "", "", false, false},
		{"empty value", "FOO=", "FOO", "", true, false},
		{"missing equals", "FOOBAR", "", "", false, true},
		{"empty key", "=value", "", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, v, ok, err := parseLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if k != tt.wantKey || v != tt.wantVal || ok != tt.wantOk {
				t.Errorf("got (%q, %q, %v), want (%q, %q, %v)",
					k, v, ok, tt.wantKey, tt.wantVal, tt.wantOk)
			}
		})
	}
}

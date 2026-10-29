package pager

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriter(t *testing.T) {
	tall := strings.Repeat("row\n", 10)
	tests := []struct {
		name  string
		cfg   Config
		input []string
		paged bool
	}{
		{"fits", Config{Width: 80, Height: 24}, []string{"a\nb\n"}, false},
		{"tall", Config{Width: 80, Height: 10}, []string{tall}, true},
		{"one short of tall", Config{Width: 80, Height: 11}, []string{tall}, false},
		{"last line without newline", Config{Width: 80, Height: 3}, []string{"a\nb\nc"}, true},
		{"wide", Config{Width: 4, Height: 24}, []string{"abcde\n"}, true},
		{"as wide as the terminal", Config{Width: 4, Height: 24}, []string{"abcd\n"}, false},
		{"wide in pieces", Config{Width: 4, Height: 24}, []string{"ab", "c", "de"}, true},
		{"wide characters", Config{Width: 5, Height: 24}, []string{"日本語\n"}, true},
		{"rune split across writes", Config{Width: 5, Height: 24}, []string{"日\xe6", "\x9c\xac語\n"}, true},
		{"escape codes take no room", Config{Width: 4, Height: 24}, []string{"\x1b[1;31mabcd\x1b[0m\n"}, false},
		{"osc takes no room", Config{Width: 4, Height: 24}, []string{"\x1b]8;;http://example.com/\x1b\\ab\x1b]8;;\x07\n"}, false},
		{"tab", Config{Width: 8, Height: 24}, []string{"a\tb\n"}, true},
		{"min lines holds back", Config{Width: 80, Height: 5, MinLines: 20}, []string{tall}, false},
		{"min lines reached", Config{Width: 80, Height: 5, MinLines: 10}, []string{tall}, true},
		{"min lines holds back wide", Config{Width: 4, Height: 24, MinLines: 2}, []string{"abcdef\n"}, false},
		{"always", Config{Always: true}, []string{"a\n"}, true},
		{"always with no output", Config{Always: true}, nil, false},
		{"size not known", Config{}, []string{tall}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, paged := sh(t, test.cfg)
			var out bytes.Buffer
			w := New(&out, cfg)
			for _, s := range test.input {
				if _, err := w.Write([]byte(s)); err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			}
			if err := w.Close(); err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			want := strings.Join(test.input, "")
			got, other := out.String(), paged()
			if test.paged {
				got, other = other, got
			}
			if got != want {
				t.Errorf("expected %q, got: %q", want, got)
			}
			if other != "" {
				t.Errorf("expected nothing on the other side, got: %q", other)
			}
		})
	}
}

func TestWriterNoCommand(t *testing.T) {
	var out bytes.Buffer
	w := New(&out, Config{Always: true})
	if _, err := w.Write([]byte("a\n")); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.String() != "a\n" {
		t.Errorf("expected the output to go straight out, got: %q", out.String())
	}
	if err := w.Close(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestWriterQuit(t *testing.T) {
	cfg, _ := sh(t, Config{Always: true})
	cfg.Command = "exit 0"
	var out bytes.Buffer
	w := New(&out, cfg)
	// enough to fill the pipe after the pager is gone
	line := []byte(strings.Repeat("x", 1023) + "\n")
	for range 1024 {
		if _, err := w.Write(line); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output, got %d bytes", out.Len())
	}
}

func TestWriterNoShell(t *testing.T) {
	cfg, _ := sh(t, Config{Width: 80, Height: 2})
	cfg.Shell = []string{filepath.Join(t.TempDir(), "missing"), "-c"}
	var out bytes.Buffer
	w := New(&out, cfg)
	if _, err := w.Write([]byte("a\nb\nc\n")); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if _, err := w.Write([]byte("d\n")); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := w.Close(); err == nil {
		t.Errorf("expected an error")
	}
	if want := "a\nb\nc\nd\n"; out.String() != want {
		t.Errorf("expected %q, got: %q", want, out.String())
	}
}

func TestWriterFailingPager(t *testing.T) {
	cfg, _ := sh(t, Config{Always: true})
	cfg.Command = "exit 3"
	w := New(new(bytes.Buffer), cfg)
	if _, err := w.Write([]byte("a\n")); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := w.Close(); err == nil || !strings.Contains(err.Error(), `running the pager "exit 3"`) {
		t.Errorf("expected the error to name the pager, got: %v", err)
	}
}

// sh adds to cfg a pager that copies its input to a file, and returns a func
// that reads the file.
func sh(t *testing.T, cfg Config) (Config, func() string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the test pager is a sh command")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	name := filepath.Join(t.TempDir(), "paged")
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	cfg.Command, cfg.Shell, cfg.Stdout, cfg.Stderr = "cat", []string{"sh", "-c"}, f, f
	return cfg, func() string {
		buf, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		return string(buf)
	}
}

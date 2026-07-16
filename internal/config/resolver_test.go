package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveReadPathPriority(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	xdg := filepath.Join(root, "xdg")
	cwd := filepath.Join(root, "cwd")

	for _, dir := range []string{home, xdg, cwd, filepath.Join(root, "empty-cwd"), filepath.Join(root, "empty-cwd-2")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	xdgPath := filepath.Join(xdg, "singctl", "config.yaml")
	homePath := filepath.Join(home, ".config", "singctl", "config.yaml")
	cwdPath := filepath.Join(cwd, ".singctl.yaml")
	explicitPath := filepath.Join(root, "custom.yaml")

	for _, path := range []string{xdgPath, homePath, cwdPath, explicitPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir parent for %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte("api:\n  token: test\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	tests := []struct {
		name string
		opts ResolveOptions
		want string
	}{
		{
			name: "explicit beats every other candidate",
			opts: ResolveOptions{
				ExplicitPath:  explicitPath,
				WorkingDir:    cwd,
				HomeDir:       home,
				XDGConfigHome: xdg,
			},
			want: explicitPath,
		},
		{
			name: "cwd beats xdg and home",
			opts: ResolveOptions{
				WorkingDir:    cwd,
				HomeDir:       home,
				XDGConfigHome: xdg,
			},
			want: cwdPath,
		},
		{
			name: "xdg beats home when cwd missing",
			opts: ResolveOptions{
				WorkingDir:    filepath.Join(root, "empty-cwd"),
				HomeDir:       home,
				XDGConfigHome: xdg,
			},
			want: xdgPath,
		},
		{
			name: "home is last fallback",
			opts: ResolveOptions{
				WorkingDir: filepath.Join(root, "empty-cwd-2"),
				HomeDir:    home,
			},
			want: homePath,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveReadPath(tt.opts)
			if err != nil {
				t.Fatalf("ResolveReadPath() error = %v", err)
			}
			if got.Path != tt.want {
				t.Fatalf("ResolveReadPath().Path = %q, want %q", got.Path, tt.want)
			}
			if !got.Exists {
				t.Fatalf("ResolveReadPath().Exists = false, want true")
			}
		})
	}
}

func TestResolveWritePathDefaultsToCanonicalLocation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	xdg := filepath.Join(root, "xdg")
	cwd := filepath.Join(root, "cwd")

	for _, dir := range []string{home, xdg, cwd} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	tests := []struct {
		name string
		opts ResolveOptions
		want string
	}{
		{
			name: "explicit path is used for writes",
			opts: ResolveOptions{
				ExplicitPath:  filepath.Join(root, "custom.yaml"),
				WorkingDir:    cwd,
				HomeDir:       home,
				XDGConfigHome: xdg,
			},
			want: filepath.Join(root, "custom.yaml"),
		},
		{
			name: "existing cwd file beats xdg",
			opts: ResolveOptions{
				WorkingDir:    cwd,
				HomeDir:       home,
				XDGConfigHome: xdg,
			},
			want: filepath.Join(cwd, ".singctl.yaml"),
		},
		{
			name: "existing cwd file is reused",
			opts: ResolveOptions{
				WorkingDir: cwd,
				HomeDir:    home,
			},
			want: filepath.Join(cwd, ".singctl.yaml"),
		},
		{
			name: "new file prefers xdg",
			opts: ResolveOptions{
				WorkingDir:    filepath.Join(root, "fresh-cwd"),
				HomeDir:       home,
				XDGConfigHome: xdg,
			},
			want: filepath.Join(xdg, "singctl", "config.yaml"),
		},
		{
			name: "new file falls back to home config",
			opts: ResolveOptions{
				WorkingDir: filepath.Join(root, "fresh-cwd-2"),
				HomeDir:    home,
			},
			want: filepath.Join(home, ".config", "singctl", "config.yaml"),
		},
	}

	cwdExisting := filepath.Join(cwd, ".singctl.yaml")
	if err := os.WriteFile(cwdExisting, []byte("output:\n  format: yaml\n"), 0o600); err != nil {
		t.Fatalf("write cwd config: %v", err)
	}
	xdgExisting := filepath.Join(xdg, "singctl", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(xdgExisting), 0o755); err != nil {
		t.Fatalf("mkdir xdg config dir: %v", err)
	}
	if err := os.WriteFile(xdgExisting, []byte("output:\n  format: table\n"), 0o600); err != nil {
		t.Fatalf("write xdg config: %v", err)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveWritePath(tt.opts)
			if err != nil {
				t.Fatalf("ResolveWritePath() error = %v", err)
			}
			if got.Path != tt.want {
				t.Fatalf("ResolveWritePath().Path = %q, want %q", got.Path, tt.want)
			}
		})
	}
}

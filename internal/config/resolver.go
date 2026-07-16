package config

import (
	"os"
	"path/filepath"
)

const (
	defaultConfigDirName  = "singctl"
	defaultConfigFileName = "config.yaml"
	localConfigFileName   = ".singctl.yaml"
)

// ResolveOptions are inputs for locating the configuration file to read or write.
type ResolveOptions struct {
	ExplicitPath  string
	WorkingDir    string
	HomeDir       string
	XDGConfigHome string
}

// ResolvedPath is a candidate config path and whether that file already exists.
type ResolvedPath struct {
	Path   string
	Exists bool
}

// ResolveReadPath picks the highest-priority existing config file, if any.
func ResolveReadPath(opts ResolveOptions) (ResolvedPath, error) {
	for _, path := range candidatePaths(opts) {
		if path == "" {
			continue
		}
		if exists(path) {
			return ResolvedPath{Path: path, Exists: true}, nil
		}
	}

	if opts.ExplicitPath != "" {
		return ResolvedPath{Path: opts.ExplicitPath, Exists: false}, nil
	}

	return ResolvedPath{Path: "", Exists: false}, nil
}

// ResolveWritePath picks where a config write should land (existing file or default create path).
func ResolveWritePath(opts ResolveOptions) (ResolvedPath, error) {
	if opts.ExplicitPath != "" {
		return ResolvedPath{
			Path:   opts.ExplicitPath,
			Exists: exists(opts.ExplicitPath),
		}, nil
	}

	readResolved, err := ResolveReadPath(opts)
	if err != nil {
		return ResolvedPath{}, err
	}
	if readResolved.Exists {
		return readResolved, nil
	}

	if opts.XDGConfigHome != "" {
		return ResolvedPath{Path: xdgConfigPath(opts.XDGConfigHome)}, nil
	}
	if opts.HomeDir != "" {
		return ResolvedPath{Path: homeConfigPath(opts.HomeDir)}, nil
	}
	return ResolvedPath{Path: filepath.Join(opts.WorkingDir, localConfigFileName)}, nil
}

func candidatePaths(opts ResolveOptions) []string {
	return []string{
		opts.ExplicitPath,
		cwdConfigPath(opts.WorkingDir),
		xdgConfigPath(opts.XDGConfigHome),
		homeConfigPath(opts.HomeDir),
	}
}

func xdgConfigPath(root string) string {
	if root == "" {
		return ""
	}
	return filepath.Join(root, defaultConfigDirName, defaultConfigFileName)
}

func homeConfigPath(home string) string {
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".config", defaultConfigDirName, defaultConfigFileName)
}

func cwdConfigPath(cwd string) string {
	if cwd == "" {
		return ""
	}
	return filepath.Join(cwd, localConfigFileName)
}

func exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

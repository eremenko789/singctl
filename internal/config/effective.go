package config

import (
	"os"
)

// ResolveOptionsFromEnv builds ResolveOptions from explicitPath and process environment.
func ResolveOptionsFromEnv(explicitPath string) (ResolveOptions, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return ResolveOptions{}, err
	}
	return ResolveOptions{
		ExplicitPath:  explicitPath,
		WorkingDir:    workingDir,
		HomeDir:       homeDir,
		XDGConfigHome: os.Getenv("XDG_CONFIG_HOME"),
	}, nil
}

// LoadEffectiveSettings resolves the config path, loads the document, and applies runtimeToken.
func LoadEffectiveSettings(explicitPath, runtimeToken string) (EffectiveSettings, error) {
	opts, err := ResolveOptionsFromEnv(explicitPath)
	if err != nil {
		return EffectiveSettings{}, err
	}
	resolved, err := ResolveReadPath(opts)
	if err != nil {
		return EffectiveSettings{}, err
	}

	cfg := DefaultConfig()
	if resolved.Exists {
		cfg, err = LoadConfig(resolved.Path)
		if err != nil {
			return EffectiveSettings{}, err
		}
	}
	if runtimeToken != "" {
		cfg.API.Token = runtimeToken
	}

	return EffectiveSettings{
		ConfigPath: resolved.Path,
		FromFile:   resolved.Exists,
		Config:     cfg,
	}, nil
}

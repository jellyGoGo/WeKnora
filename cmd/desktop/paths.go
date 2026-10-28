package main

import (
	"os"
	"path/filepath"
	"strings"
)

func desktopResourceDir(execPath, cwd string) string {
	var candidates []string
	if execPath != "" {
		if strings.Contains(filepath.ToSlash(execPath), ".app/Contents/MacOS") {
			candidates = append(candidates, filepath.Join(filepath.Dir(filepath.Dir(execPath)), "Resources"))
		}
		candidates = append(candidates, filepath.Dir(execPath))
	}
	candidates = append(candidates, cwd, filepath.Clean(filepath.Join(cwd, "..", "..")))
	for _, dir := range candidates {
		if info, err := os.Stat(filepath.Join(dir, "config", "config.yaml")); err == nil && !info.IsDir() {
			return dir
		}
	}
	return cwd
}

func configureWindowsDesktopStorage(appDataDir string) error {
	dbPath := resolveDesktopDataPath(os.Getenv("DB_PATH"), filepath.Join("data", "weknora.db"), appDataDir)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}
	if err := os.Setenv("DB_PATH", dbPath); err != nil {
		return err
	}
	if strings.TrimSpace(os.Getenv("LOG_PATH")) == "" {
		return os.Setenv("LOG_PATH", filepath.Join(appDataDir, "logs", "WeKnora Lite.log"))
	}
	return nil
}

func resolveDesktopDataPath(rawPath, defaultRelativePath, appSupportDir string) string {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		trimmed = defaultRelativePath
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed)
	}
	trimmed = strings.TrimPrefix(trimmed, "."+string(filepath.Separator))
	return filepath.Join(appSupportDir, filepath.Clean(trimmed))
}

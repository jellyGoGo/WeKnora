package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDesktopResourceDirFindsInstalledResources(t *testing.T) {
	appDir := filepath.Join(t.TempDir(), "WeKnora Lite")
	if err := os.MkdirAll(filepath.Join(appDir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "config", "config.yaml"), []byte("server: {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd := t.TempDir()
	if got := desktopResourceDir(filepath.Join(appDir, "WeKnora Lite.exe"), cwd); got != appDir {
		t.Fatalf("resource directory = %q, want %q", got, appDir)
	}
	if got := desktopResourceDir("", cwd); got != cwd {
		t.Fatalf("missing bundle should preserve working directory, got %q", got)
	}
}

func TestDesktopResourceDirPreservesMacBundleAndDevelopment(t *testing.T) {
	root := t.TempDir()
	resources := filepath.Join(root, "WeKnora.app", "Contents", "Resources")
	if err := os.MkdirAll(filepath.Join(resources, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(resources, "config", "config.yaml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(root, "WeKnora.app", "Contents", "MacOS", "WeKnora")
	if got := desktopResourceDir(executable, root); got != resources {
		t.Fatalf("macOS resources = %q, want %q", got, resources)
	}
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "config.yaml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := desktopResourceDir("", filepath.Join(root, "cmd", "desktop")); got != root {
		t.Fatalf("development resources = %q, want %q", got, root)
	}
}

func TestWindowsDesktopStorageUsesUserDirectory(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "WeKnora Lite")
	t.Setenv("DB_PATH", "./data/weknora.db")
	t.Setenv("LOG_PATH", "")
	if err := configureWindowsDesktopStorage(dataDir); err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dataDir, "data", "weknora.db"); os.Getenv("DB_PATH") != want {
		t.Fatalf("database path = %q, want %q", os.Getenv("DB_PATH"), want)
	}
	if want := filepath.Join(dataDir, "logs", "WeKnora Lite.log"); os.Getenv("LOG_PATH") != want {
		t.Fatalf("log path = %q, want %q", os.Getenv("LOG_PATH"), want)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "data")); err != nil {
		t.Fatal(err)
	}
	customDB := filepath.Join(t.TempDir(), "custom.db")
	customLog := filepath.Join(t.TempDir(), "custom.log")
	t.Setenv("DB_PATH", customDB)
	t.Setenv("LOG_PATH", customLog)
	if err := configureWindowsDesktopStorage(dataDir); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("DB_PATH") != customDB || os.Getenv("LOG_PATH") != customLog {
		t.Fatal("explicit absolute database and log paths must be preserved")
	}
}

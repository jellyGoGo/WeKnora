package types

import (
	"os"
	"path/filepath"
)

// Package initialization runs before main can load .env or change directory.
// Resolve installed dictionaries from the executable, not the build machine.
func resolveJiebaDictDir(execPath, configured string) string {
	if configured != "" {
		return configured
	}
	if execPath == "" {
		return ""
	}
	dir := filepath.Join(filepath.Dir(execPath), "dict")
	for _, name := range []string{"jieba.dict.utf8", "hmm_model.utf8", "user.dict.utf8", "idf.utf8", "stop_words.utf8"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || info.IsDir() {
			return ""
		}
	}
	return dir
}

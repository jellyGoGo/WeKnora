package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveJiebaDictDirUsesInstalledDictionaries(t *testing.T) {
	appDir := filepath.Join(t.TempDir(), "WeKnora Lite")
	dictDir := filepath.Join(appDir, "dict")
	if err := os.MkdirAll(dictDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"jieba.dict.utf8", "hmm_model.utf8", "user.dict.utf8", "idf.utf8", "stop_words.utf8"} {
		if err := os.WriteFile(filepath.Join(dictDir, name), []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(appDir, "WeKnora Lite.exe")
	if got := resolveJiebaDictDir(executable, ""); got != dictDir {
		t.Fatalf("installed dictionary directory = %q, want %q", got, dictDir)
	}
	if got := resolveJiebaDictDir(executable, "custom-dictionaries"); got != "custom-dictionaries" {
		t.Fatalf("explicit dictionary directory was replaced: %q", got)
	}
	if err := os.Remove(filepath.Join(dictDir, "hmm_model.utf8")); err != nil {
		t.Fatal(err)
	}
	if got := resolveJiebaDictDir(executable, ""); got != "" {
		t.Fatalf("incomplete bundle should fall back to gojieba defaults, got %q", got)
	}
}

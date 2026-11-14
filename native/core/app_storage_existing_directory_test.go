package core

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestNativeVerifyEmptyDownloadsTarget(t *testing.T) {
    temp := t.TempDir()
    selected := filepath.Join(temp, "selected")
    if err := os.Mkdir(selected, 0755); err != nil { t.Fatal(err) }
    if err := nativeVerifyEmptyDownloadsTarget(selected); err != nil {
        t.Fatalf("Android folder picker provides an existing empty folder: %v", err)
    }
    if err := nativeVerifyEmptyDownloadsTarget(filepath.Join(temp, "not-yet-created")); err != nil {
        t.Fatalf("internal default folder may be created later: %v", err)
    }
    if err := os.WriteFile(filepath.Join(selected, "personal-file.txt"), []byte("keep"), 0600); err != nil { t.Fatal(err) }
    if err := nativeVerifyEmptyDownloadsTarget(selected); err == nil ||
        !strings.Contains(err.Error(), "空") {
        t.Fatalf("must reject non-empty selected folder without overwriting data: %v", err)
    }
    if got, err := os.ReadFile(filepath.Join(selected, "personal-file.txt")); err != nil || string(got) != "keep" {
        t.Fatalf("must preserve unrelated content: %v %q", err, got)
    }
    if err := nativeVerifyEmptyDownloadsTarget(filepath.Join(selected, "personal-file.txt")); err == nil {
        t.Fatal("a file is not a valid download directory")
    }
}

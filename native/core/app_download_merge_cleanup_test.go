package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergedCleanupRemovesOnlySelectedFilesAndGeneratedFolders(t *testing.T) {
	manager := downloadTestManager(t)
	drama := nativeDrama{ID: "hongguo:990", Source: "hongguo", Title: "清理测试"}
	one := convertingFixture(t, manager, drama, 1, false)
	two := convertingFixture(t, manager, drama, 2, false)
	selectedOutput := convertFixture(t, manager, one)
	otherOutput := convertFixture(t, manager, two)
	raw := manager.jobDirectory(manager.jobs[one].nativeDownloadJob)
	os.MkdirAll(raw, 0700)
	for _, name := range []string{"media.mp4", "index.m3u8", "key.bin", "segment.part", ".bundle"} {
		os.WriteFile(filepath.Join(raw, name), []byte("generated"), 0600)
	}
	merged := filepath.Join(manager.root, "清理测试-合.mp4")
	os.WriteFile(merged, []byte("verified-output"), 0600)
	manager.engine.workLease("media", "start")
	defer manager.engine.workLease("media", "end")
	expected := map[string]string{one: nativeMergedSourceVersion(manager.jobs[one].nativeDownloadJob)}
	result, err := manager.cleanupMergedDownloads([]string{one}, expected, merged)
	if err != nil || len(result.Completed) != 1 {
		t.Fatalf("cleanup: %+v %v", result, err)
	}
	for _, name := range []string{raw, manager.outputFile(selectedOutput)} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Fatal("selected download files remained", name)
		}
	}
	for _, name := range []string{merged, manager.outputFile(otherOutput), filepath.Join(manager.root, "index.json")} {
		if _, err := os.Stat(name); err != nil {
			t.Fatal("unrelated or merged file deleted", name, err)
		}
	}
	if len(manager.jobs) != 1 || manager.jobs[two] == nil {
		t.Fatal("unselected index entry deleted")
	}
	expected = map[string]string{two: nativeMergedSourceVersion(manager.jobs[two].nativeDownloadJob)}
	if _, err := manager.cleanupMergedDownloads([]string{two}, expected, merged); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(manager.outputFile(otherOutput))); !os.IsNotExist(err) {
		t.Fatal("empty drama directory remained")
	}
}

func TestMergedCleanupRejectsChangedVersionsAndMissingOutput(t *testing.T) {
	manager := downloadTestManager(t)
	drama := nativeDrama{ID: "hongguo:991", Source: "hongguo", Title: "保留测试"}
	id := convertingFixture(t, manager, drama, 1, false)
	output := convertFixture(t, manager, id)
	expected := map[string]string{id: nativeMergedSourceVersion(manager.jobs[id].nativeDownloadJob)}
	merged := filepath.Join(manager.root, "成品.mp4")
	manager.engine.workLease("media", "start")
	defer manager.engine.workLease("media", "end")
	if _, err := manager.cleanupMergedDownloads([]string{id}, expected, merged); err == nil {
		t.Fatal("deleted before merged output existed")
	}
	os.WriteFile(merged, []byte("output"), 0600)
	manager.jobs[id].Revision++
	if _, err := manager.cleanupMergedDownloads([]string{id}, expected, merged); err == nil {
		t.Fatal("changed source deleted")
	}
	if _, err := os.Stat(manager.outputFile(output)); err != nil {
		t.Fatal("source removed on validation failure")
	}
}

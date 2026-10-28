package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadDirectorySwitchRetainsAndRestoresOldFiles(t *testing.T) {
	engine, id := storageFixture(t)
	old := engine.downloads.root
	next := t.TempDir()
	if err := engine.changeDownloadDirectory(next); err != nil {
		t.Fatal(err)
	}
	if len(engine.downloads.jobs) != 0 {
		t.Fatal("old jobs leaked into new directory")
	}
	if _, err := os.Stat(filepath.Join(old, id, "media.mp4")); err != nil {
		t.Fatal("old media removed", err)
	}
	if nativeDownloadLocation(engine.directory) != next {
		t.Fatal("new directory not persisted")
	}
	if err := engine.changeDownloadDirectory(old); err != nil {
		t.Fatal(err)
	}
	if len(engine.downloads.jobs) != 1 {
		t.Fatal("old download index not restored")
	}
}

func TestDownloadDirectoryRejectsForeignIndexAndActiveTasks(t *testing.T) {
	engine, id := storageFixture(t)
	old := engine.downloads.root
	foreign := t.TempDir()
	os.WriteFile(filepath.Join(foreign, "index.json"), []byte("[]"), 0600)
	if err := engine.changeDownloadDirectory(foreign); err == nil {
		t.Fatal("foreign index accepted")
	}
	if engine.downloads.root != old {
		t.Fatal("failed switch changed root")
	}
	engine.downloads.jobs[id].State = "queued"
	if err := engine.changeDownloadDirectory(t.TempDir()); err == nil {
		t.Fatal("queued task ignored")
	}
	engine.downloads.jobs[id].State = "completed"
	if _, err := engine.workLease("media", "start"); err != nil {
		t.Fatal(err)
	}
	if err := engine.changeDownloadDirectory(t.TempDir()); err == nil {
		t.Fatal("switch during merge")
	}
	if err := engine.downloads.control(id, "remove"); err == nil {
		t.Fatal("removed merging source")
	}
	engine.workLease("media", "end")
}

func TestDownloadReadableFolderCollisionAndRestart(t *testing.T) {
	manager := downloadTestManager(t)
	manager.resolve = func(ctx context.Context, job nativeDownloadJob) (providerMedia, error) {
		<-ctx.Done()
		return providerMedia{}, ctx.Err()
	}
	input := downloadTestInput(2)
	input.Drama.Title = "剧:名/测试"
	if _, err := manager.enqueue(input); err != nil {
		t.Fatal(err)
	}
	one := manager.jobs[nativeDownloadID(input.Drama.ID, 1)]
	two := manager.jobs[nativeDownloadID(input.Drama.ID, 2)]
	if one.Folder != "剧_名_测试/0001" || two.Folder != "剧_名_测试/0002" {
		t.Fatalf("wrong episode layout: %s %s", one.Folder, two.Folder)
	}
	input.Drama.ID = "hongguo:200"
	if _, err := manager.enqueue(input); err != nil {
		t.Fatal(err)
	}
	other := manager.jobs[nativeDownloadID(input.Drama.ID, 1)]
	if other.Folder != "剧_名_测试 (2)/0001" {
		t.Fatalf("same-title collision: %s", other.Folder)
	}
	manager.close()
	reopened := newNativeDownloads(manager.engine)
	defer reopened.close()
	if reopened.loadErr != nil || len(reopened.jobs) != 4 {
		t.Fatalf("restart: %v", reopened.loadErr)
	}
}

func TestConvertedReadableOutputRemovesRawAndKeyAfterSuccess(t *testing.T) {
	manager := downloadTestManager(t)
	drama := nativeDrama{ID: "hongguo:900", Source: "hongguo", Title: "样本"}
	id := convertingFixture(t, manager, drama, 1, false)
	record := manager.jobs[id]
	raw := manager.jobDirectory(record.nativeDownloadJob)
	targetRaw := filepath.Join(manager.root, "样本", "0001")
	os.MkdirAll(filepath.Dir(targetRaw), 0700)
	if err := os.Rename(raw, targetRaw); err != nil {
		t.Fatal(err)
	}
	record.Folder = "样本/0001"
	output := convertFixture(t, manager, id)
	if output != "样本/第1集.mp4" {
		t.Fatal(output)
	}
	if _, err := os.Stat(targetRaw); !os.IsNotExist(err) {
		t.Fatal("raw episode folder kept", err)
	}
	if record.File != "" || record.Key != "" {
		t.Fatal("raw metadata and key kept after conversion")
	}
	plan, found, err := manager.localPlan(drama.ID, 1)
	if err != nil || !found || !plan.Local || plan.Key != "" {
		t.Fatalf("plaintext MP4 was decrypted again: %+v %v", plan, err)
	}
}

func TestConvertedOutputKeepsRawWhenIndexSaveFails(t *testing.T) {
	manager := downloadTestManager(t)
	drama := nativeDrama{ID: "hongguo:901", Source: "hongguo", Title: "保存失败测试"}
	id := convertingFixture(t, manager, drama, 1, false)
	work, err := manager.beginConversion(id)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(work.Output, []byte("synthetic-video"), 0600)
	manager.loadErr = errors.New("synthetic index save failure")
	if err := manager.finishConversion(id, "done"); err == nil {
		t.Fatal("index save failure ignored")
	}
	manager.loadErr = nil
	if _, err := os.Stat(work.Input); err != nil {
		t.Fatal("raw lost after save failure")
	}
	if manager.jobs[id].File == "" || manager.jobs[id].Key == "" {
		t.Fatal("retry metadata lost")
	}
	convertFixture(t, manager, id)
}

func TestDownloadReadableLayoutReusesLegacyMP4Folder(t *testing.T) {
	manager := downloadTestManager(t)
	drama := nativeDrama{ID: "hongguo:902", Source: "hongguo", Title: "旧剧名"}
	id := convertingFixture(t, manager, drama, 1, false)
	output := convertFixture(t, manager, id)
	manager.resolve = func(ctx context.Context, job nativeDownloadJob) (providerMedia, error) {
		<-ctx.Done()
		return providerMedia{}, ctx.Err()
	}
	input := downloadTestInput(1)
	input.Drama = drama
	input.Entries[0].Index = 2
	if _, err := manager.enqueue(input); err != nil {
		t.Fatal(err)
	}
	job := manager.jobs[nativeDownloadID(drama.ID, 2)]
	if filepath.ToSlash(filepath.Dir(output))+"/0002" != job.Folder {
		t.Fatalf("legacy drama was split into a new folder: %s", job.Folder)
	}
}

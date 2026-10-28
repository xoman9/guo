package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func nativeMergedSourceVersion(job nativeDownloadJob) string {
	return fmt.Sprintf("%d-%d-%d-%d", job.Created, job.Revision, job.Bytes, job.ActualQuality)
}

func (manager *nativeDownloads) cleanupMergedDownloads(ids []string, expected map[string]string, output string) (nativeDownloadControlResult, error) {
	result := nativeDownloadControlResult{Completed: []string{}, Failures: map[string]string{}}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if !manager.mediaBusy || manager.closed || manager.moving {
		return result, errors.New("请在合并任务中清理原分集")
	}
	if manager.loadErr != nil {
		return result, manager.loadErr
	}
	if len(ids) == 0 || len(ids) > 5000 {
		return result, errors.New("请选择有效的合并分集")
	}
	output = filepath.Clean(output)
	if filepath.Dir(output) != manager.root || !strings.HasSuffix(strings.ToLower(output), ".mp4") {
		return result, errors.New("合并成品路径无效，原分集已保留")
	}
	info, err := os.Lstat(output)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 {
		return result, errors.New("合并成品未保存，原分集已保留")
	}
	selected := map[string]*nativeDownloadRecord{}
	for _, id := range ids {
		job := manager.jobs[id]
		if job == nil || job.State != "completed" || expected[id] == "" || nativeMergedSourceVersion(job.nativeDownloadJob) != expected[id] || manager.active[id] != nil || manager.converting[id] {
			return result, errors.New("原分集已变更，已停止清理并保留文件")
		}
		if !nativeDownloadAvailable(job.nativeDownloadJob) {
			return result, errNativeBuildSource
		}
		selected[id] = job
	}
	for _, job := range selected {
		job.State = "removing"
	}
	if err := manager.saveLocked(); err != nil {
		for _, job := range selected {
			job.State = "completed"
		}
		return result, err
	}
	for id, job := range selected {
		if err := manager.removeFiles(*job); err != nil {
			job.Error = "成品已保存，原分集清理失败，请重试删除"
			result.Failures[id] = job.Error
			continue
		}
		for _, parent := range []string{filepath.Dir(manager.jobDirectory(job.nativeDownloadJob)), filepath.Dir(manager.outputFile(job.Output))} {
			for parent != manager.root {
				relative, err := filepath.Rel(manager.root, parent)
				if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
					break
				}
				if err := os.Remove(parent); err != nil && !os.IsNotExist(err) {
					break
				}
				next := filepath.Dir(parent)
				if next == parent {
					break
				}
				parent = next
			}
		}
		delete(manager.jobs, id)
		result.Completed = append(result.Completed, id)
	}
	if err := manager.saveLocked(); err != nil {
		return result, err
	}
	return result, nil
}

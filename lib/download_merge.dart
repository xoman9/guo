import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:path/path.dart' as path;

import 'core_bridge.dart';
import 'media_pipeline.dart';
import 'models.dart';

List<DownloadJob> continuousDownloadJobs(List<DownloadJob> selected) {
  final jobs = [...selected]
    ..sort((a, b) => a.episode.number.compareTo(b.episode.number));
  if (jobs.isEmpty ||
      jobs.any((job) => !job.completed) ||
      jobs.map((job) => job.drama.id).toSet().length != 1) {
    throw AppFailure('请选择同一部剧已完成下载的分集');
  }
  for (var index = 1; index < jobs.length; index++) {
    if (jobs[index].episode.number != jobs[index - 1].episode.number + 1) {
      throw AppFailure('分集必须连续且不能重复，请补齐缺集后合并');
    }
  }
  return jobs;
}

String downloadMergeName(String title) {
  var name = String.fromCharCodes(title.runes.take(60))
      .replaceAll(RegExp(r'[\x00-\x1f\x7f<>:"/\\|?*]'), '_')
      .replaceAll(RegExp(r'^[\s.]+|[\s.]+$'), '');
  if (name.isEmpty) name = '未命名';
  if (RegExp(
    r'^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(?:\.|$)',
    caseSensitive: false,
  ).hasMatch(name)) {
    name = '_$name';
  }
  return '$name-合';
}

class DownloadMerger extends ChangeNotifier {
  DownloadMerger(this.repository, {MediaExecutor? executor})
    : executor = executor ?? FFmpegExecutor();
  final AppRepository repository;
  final MediaExecutor executor;
  String status = '';
  double progress = 0;
  bool busy = false, _cancelled = false, _disposed = false, _cleaning = false;
  bool get cancellable => busy && !_cleaning;

  void _check() {
    if (_cancelled) throw AppFailure('已取消合并，原分集已保留');
  }

  void _update(String message, double value) {
    status = message;
    progress = value;
    if (!_disposed) notifyListeners();
  }

  Future<void> cancel() async {
    if (_cleaning) return;
    _cancelled = true;
    await executor.cancel();
  }

  Future<String> merge(List<DownloadJob> selected) async {
    if (busy) throw AppFailure('已有影片正在合并');
    final jobs = continuousDownloadJobs(selected);
    busy = true;
    _cancelled = false;
    Directory? workspace;
    File? reserved;
    var leased = false, completed = false;
    try {
      _update('检查已下载分集', 0);
      await repository.workLease('media', 'start');
      leased = true;
      final current = {
        for (final job in await repository.downloads()) job.id: job,
      };
      for (final job in jobs) {
        final latest = current[job.id];
        if (latest == null ||
            !latest.completed ||
            latest.created != job.created ||
            latest.revision != job.revision ||
            latest.bytes != job.bytes) {
          throw AppFailure('下载记录已改变，请刷新后重新选择');
        }
      }
      final root = await repository.downloadDirectory();
      if (!path.isAbsolute(root)) throw AppFailure('下载资料夹无效');
      workspace = await Directory(root).createTemp('.guo-merge-');
      final files = <String>[];
      var totalDuration = 0.0;
      Map<String, dynamic>? firstVideo;
      for (var index = 0; index < jobs.length; index++) {
        _check();
        final job = jobs[index];
        final plan = await repository.resolve(job.drama, job.episode);
        if (!plan.local || !await File(plan.url).exists()) {
          throw AppFailure('第 ${job.episode.number} 集没有完整本地影片，请重新下载');
        }
        final probe = await executor.probe(plan.url);
        verifyMediaDuration(probe, 0);
        final first = firstVideo;
        if (first != null &&
            (first['codec_name'] != probe.video['codec_name'] ||
                first['width'] != probe.video['width'] ||
                first['height'] != probe.video['height'])) {
          throw AppFailure('分集编码或尺寸不同，请使用相同画质重新下载后合并');
        }
        firstVideo ??= probe.video;
        totalDuration += probe.duration;
        final output = path.join(workspace.path, 'episode-$index.ts');
        _update('准备第 ${job.episode.number} 集', index / jobs.length * .7);
        await executor.run(
          [
            if (plan.decryptionKey.isNotEmpty) ...[
              '-decryption_key',
              plan.decryptionKey,
            ],
            if (path.extension(plan.url).toLowerCase() == '.m3u8') ...[
              '-allowed_extensions',
              'ALL',
              '-extension_picky',
              '0',
            ],
            '-i',
            plan.url,
            '-map',
            '0:v:0',
            '-map',
            '0:a:0?',
            '-c',
            'copy',
            '-f',
            'mpegts',
            output,
          ],
          duration: probe.duration,
          progress: (value) => _update(
            '准备第 ${job.episode.number} 集',
            (index + value) / jobs.length * .7,
          ),
        );
        files.add(path.basename(output));
      }
      _check();
      final playlist = File(path.join(workspace.path, 'concat.txt'));
      await playlist.writeAsString(
        files.map((file) => "file '$file'").join('\n'),
        flush: true,
      );
      final merged = File(path.join(workspace.path, 'merged.mp4'));
      _update('合并影片', .7);
      await executor.run(
        [
          '-f',
          'concat',
          '-safe',
          '1',
          '-i',
          playlist.path,
          '-map',
          '0:v:0',
          '-map',
          '0:a:0?',
          '-c',
          'copy',
          '-movflags',
          '+faststart',
          merged.path,
        ],
        duration: totalDuration,
        progress: (value) => _update('合并影片', .7 + value * .15),
      );
      _check();
      verifyMediaDuration(await executor.probe(merged.path), totalDuration);
      _update('验证成品', .85);
      await executor.run(
        [
          '-xerror',
          '-err_detect',
          'explode',
          '-i',
          merged.path,
          '-map',
          '0:v:0',
          '-map',
          '0:a:0?',
          '-f',
          'null',
          '-',
        ],
        duration: totalDuration,
        progress: (value) => _update('验证成品', .85 + value * .1),
      );
      _check();
      final base = downloadMergeName(jobs.first.drama.title);
      for (var index = 0; index < 10000; index++) {
        final file = File(
          path.join(root, index == 0 ? '$base.mp4' : '$base ($index).mp4'),
        );
        try {
          await file.create(exclusive: true);
          reserved = file;
          break;
        } on FileSystemException {
          if (!await file.exists()) rethrow;
        }
      }
      final target = reserved;
      if (target == null) throw AppFailure('同名合并成品过多');
      _update('保存合并成品', .95);
      await merged.openRead().pipe(target.openWrite());
      _check();
      if (await target.length() != await merged.length())
        throw AppFailure('保存不完整，原分集已保留');
      final saved = await target.open(mode: FileMode.append);
      try {
        await saved.flush();
      } finally {
        await saved.close();
      }
      _check();
      completed = true;
      _cleaning = true;
      _update('成品已保存，正在清理原分集与下载文件', .98);
      try {
        await repository.cleanupMergedDownloads(jobs, target.path);
      } catch (error) {
        throw AppFailure('合并成品已保存于 ${target.path}；清理未完成：$error');
      }
      _update('合并完成，原分集与下载文件已清理', 1);
      return target.path;
    } finally {
      try {
        if (!completed && reserved != null && await reserved.exists())
          await reserved.delete();
        if (workspace != null && await workspace.exists())
          await workspace.delete(recursive: true);
      } finally {
        try {
          if (leased) await repository.workLease('media', 'end');
        } finally {
          _cleaning = false;
          busy = false;
          if (!_disposed) notifyListeners();
        }
      }
    }
  }

  @override
  void dispose() {
    _disposed = true;
    if (busy) cancel();
    super.dispose();
  }
}

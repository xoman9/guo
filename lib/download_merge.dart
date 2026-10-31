import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:path/path.dart' as path;

import 'core_bridge.dart';
import 'hardware_video_encoder.dart';
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

String downloadMergeName(String title, int firstEpisode, int lastEpisode) {
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
  return '$name($firstEpisode-$lastEpisode)';
}

class DownloadMerger extends ChangeNotifier {
  DownloadMerger(
    this.repository, {
    MediaExecutor? executor,
    this.hardwareEncoderLookup = findAndroidAvcEncoder,
  }) : executor = executor ?? FFmpegExecutor();
  final AppRepository repository;
  final MediaExecutor executor;
  final Future<AvcHardwareEncoder?> Function(MergePlan) hardwareEncoderLookup;
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
      final prepared = <String>[];
      final probes = <MediaProbe>[];
      for (var index = 0; index < jobs.length; index++) {
        _check();
        final job = jobs[index];
        _update('預備第 ${job.episode.number} 集', index / jobs.length * .2);
        final local = await repository.resolve(job.drama, job.episode);
        _check();
        if (!local.local || !await File(local.url).exists()) {
          throw AppFailure('第 ${job.episode.number} 集没有完整本地影片，请重新下载');
        }
        if (local.decryptionKey.isEmpty &&
            path.extension(local.url).toLowerCase() != '.m3u8') {
          probes.add(await executor.probe(local.url));
          prepared.add(local.url);
          continue;
        }
        final output = path.join(workspace.path, 'original-$index.mkv');
        await executor.run([
          if (local.decryptionKey.isNotEmpty) ...[
            '-decryption_key',
            local.decryptionKey,
          ],
          if (path.extension(local.url).toLowerCase() == '.m3u8') ...[
            '-allowed_extensions',
            'ALL',
            '-extension_picky',
            '0',
          ],
          '-i',
          local.url,
          '-map',
          '0:v:0',
          '-map',
          '0:a:0?',
          '-c',
          'copy',
          '-map_metadata',
          '-1',
          '-avoid_negative_ts',
          'make_zero',
          output,
        ]);
        _check();
        probes.add(await executor.probe(output));
        prepared.add(output);
      }
      _check();
      final plan = MergePlan.create(probes);
      final totalDuration = probes.fold<double>(
        0,
        (sum, probe) => sum + probe.duration,
      );
      final hardware = await hardwareEncoderLookup(plan);
      _check();
      var hardwareFailed = false;
      final files = <String>[];
      for (var index = 0; index < jobs.length; index++) {
        _check();
        final accelerated =
            hardware != null && !hardwareFailed && plan.videoChanges[index];
        final message = accelerated
            ? '第 ${jobs[index].episode.number} 集 H.264 硬體轉檔中'
            : plan.videoChanges[index] || plan.audioChanges[index]
            ? '第 ${jobs[index].episode.number} 集格式不符，CPU 轉檔中'
            : '預備第 ${jobs[index].episode.number} 集';
        final output = path.join(
          workspace.path,
          'episode-$index.${plan.transportStream ? 'ts' : 'mkv'}',
        );
        _update(message, .2 + index / jobs.length * .5);
        try {
          await executor.run(
            plan.normalizeArguments(
              prepared[index],
              output,
              probes[index],
              index,
              hardware: accelerated ? hardware : null,
            ),
            duration: probes[index].duration,
            progress: (value) =>
                _update(message, .2 + (index + value) / jobs.length * .5),
          );
        } catch (_) {
          _check();
          if (!accelerated) rethrow;
          hardwareFailed = true;
          final fallback = '第 ${jobs[index].episode.number} 集硬體轉檔不可用，改用 CPU 轉檔';
          _update(fallback, .2 + index / jobs.length * .5);
          await executor.run(
            plan.normalizeArguments(
              prepared[index],
              output,
              probes[index],
              index,
            ),
            duration: probes[index].duration,
            progress: (value) =>
                _update(fallback, .2 + (index + value) / jobs.length * .5),
          );
        }
        files.add(path.basename(output));
      }
      _check();
      final playlist = File(path.join(workspace.path, 'concat.txt'));
      await playlist.writeAsString(
        files.map((file) => "file '$file'").join('\n'),
        flush: true,
      );
      final merged = File(path.join(workspace.path, 'merged.mp4'));
      _update('合併第 ${jobs.first.episode.number} 集中', .7);
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
        progress: (value) {
          final elapsed = value * totalDuration;
          var boundary = 0.0;
          var index = jobs.length - 1;
          for (var i = 0; i < probes.length; i++) {
            boundary += probes[i].duration;
            if (elapsed < boundary) {
              index = i;
              break;
            }
          }
          _update('合併第 ${jobs[index].episode.number} 集中', .7 + value * .25);
        },
      );
      _check();
      final base = downloadMergeName(
        jobs.first.drama.title,
        jobs.first.episode.number,
        jobs.last.episode.number,
      );
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

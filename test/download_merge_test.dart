import 'dart:io';
import 'dart:convert';

import 'package:duanju_app/core_bridge.dart';
import 'package:duanju_app/download_merge.dart';
import 'package:duanju_app/media_pipeline.dart';
import 'package:duanju_app/models.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:path/path.dart' as path;

import 'support/fakes.dart';

class _MergeRepository extends FakeRepository {
  _MergeRepository(this.directory);
  final Directory directory;
  final leases = <String>[];
  bool failCleanup = false;
  @override
  Future<void> cleanupMergedDownloads(
    List<DownloadJob> selected,
    String output,
  ) async {
    if (failCleanup) throw AppFailure('synthetic cleanup failure');
    if (!await File(output).exists()) throw AppFailure('missing merged output');
    for (final job in selected) {
      await File(path.join(directory.path, '${job.episode.number}.mp4'))
          .delete();
    }
    final removed = selected.map((job) => job.id).toSet();
    jobs = jobs!.where((job) => !removed.contains(job.id)).toList();
  }

  @override
  Future<String> downloadDirectory() async => directory.path;
  @override
  Future<int> workLease(String id, String command) async {
    leases.add('$id:$command');
    return command == 'start' ? 1 : 0;
  }

  @override
  Future<PlaybackPlan> resolve(
    Drama drama,
    Episode episode, {
    int quality = 0,
  }) async => PlaybackPlan(
    url: path.join(directory.path, '${episode.number}.mp4'),
    local: true,
  );
}

class _MergeExecutor implements MediaExecutor {
  bool fail = false, mixed = false;
  final commands = <List<String>>[];
  final probed = <String>[];
  @override
  Future<MediaProbe> probe(String file) async {
    probed.add(file);
    return MediaProbe({
      'format': {'duration': file.endsWith('merged.mp4') ? '24' : '12'},
      'streams': [
        {
          'codec_type': 'video',
          'codec_name': 'h264',
          'width':
              mixed &&
                  (file.endsWith('original-1.mkv') ||
                      path.basename(file) == '2.mp4')
              ? 360
              : 720,
          'height':
              mixed &&
                  (file.endsWith('original-1.mkv') ||
                      path.basename(file) == '2.mp4')
              ? 640
              : 1280,
          'pix_fmt': 'yuv420p',
        },
      ],
    });
  }

  @override
  Future<void> run(
    List<String> arguments, {
    double duration = 0,
    void Function(double)? progress,
  }) async {
    commands.add(List.of(arguments));
    if (arguments.contains('libx264') && probed.length != 2)
      throw StateError('all episodes must be probed before conversion');
    if (fail) throw AppFailure('synthetic media failure');
    if (arguments.last != '-')
      await File(arguments.last).writeAsString('synthetic-video');
    progress?.call(1);
  }

  @override
  Future<void> cancel() async {}
}

class _ProcessMergeExecutor implements MediaExecutor {
  final commands = <List<String>>[];
  @override
  Future<MediaProbe> probe(String file) async {
    final result = await Process.run('ffprobe', [
      '-v',
      'error',
      '-show_format',
      '-show_streams',
      '-show_data_hash',
      'sha256',
      '-of',
      'json',
      file,
    ]);
    if (result.exitCode != 0) throw StateError('${result.stderr}');
    return MediaProbe(
      Map<String, dynamic>.from(jsonDecode(result.stdout as String) as Map),
    );
  }

  @override
  Future<void> run(
    List<String> arguments, {
    double duration = 0,
    void Function(double)? progress,
  }) async {
    commands.add(List.of(arguments));
    final result = await Process.run('ffmpeg', [
      '-v',
      'error',
      '-nostdin',
      '-y',
      ...arguments,
    ]);
    if (result.exitCode != 0) throw StateError('${result.stderr}');
    progress?.call(1);
  }

  @override
  Future<void> cancel() async {}
}

void main() {
  test('merge selection must be completed, same drama and consecutive', () {
    final show = drama('a');
    expect(() => continuousDownloadJobs([]), throwsA(isA<AppFailure>()));
    expect(
      () =>
          continuousDownloadJobs([downloadJob(show, 1), downloadJob(show, 3)]),
      throwsA(isA<AppFailure>()),
    );
    expect(
      () =>
          continuousDownloadJobs([downloadJob(show, 1), downloadJob(show, 1)]),
      throwsA(isA<AppFailure>()),
    );
    expect(
      () => continuousDownloadJobs([
        downloadJob(show, 1),
        downloadJob(drama('b'), 2),
      ]),
      throwsA(isA<AppFailure>()),
    );
    expect(
      () => continuousDownloadJobs([downloadJob(show, 1, state: 'paused')]),
      throwsA(isA<AppFailure>()),
    );
    expect(
      continuousDownloadJobs([downloadJob(show, 2), downloadJob(show, 1)])
          .first
          .episode
          .number,
      1,
    );
  });

  test('merge names are safe on Android and Windows', () {
    expect(downloadMergeName('  甲:剧/名?  ', 1, 200), '甲_剧_名_(1-200)');
    expect(downloadMergeName('..', 1, 1), '未命名(1-1)');
    expect(downloadMergeName('CON', 5, 8), '_CON(5-8)');
  });

  test('merge normalizes lowest resolution without verification then removes originals', () async {
    final directory = await Directory.systemTemp.createTemp('guo-merge-test-');
    addTearDown(() => directory.delete(recursive: true));
    final repository = _MergeRepository(directory);
    final show = drama('a', title: '样本');
    repository.jobs = [downloadJob(show, 1), downloadJob(show, 2)];
    for (var number = 1; number <= 2; number++) {
      await File(path.join(directory.path, '$number.mp4'))
          .writeAsString('source-$number');
    }
    final existing = File(path.join(directory.path, '样本(1-2).mp4'));
    await existing.writeAsString('keep-existing');
    final executor = _MergeExecutor()..mixed = true;
    final merger = DownloadMerger(repository, executor: executor);
    final statuses = <String>[];
    merger.addListener(() => statuses.add(merger.status));
    addTearDown(merger.dispose);
    final output = await merger.merge(repository.jobs!);
    expect(path.basename(output), '样本(1-2) (1).mp4');
    expect(await existing.readAsString(), 'keep-existing');
    expect(await File(path.join(directory.path, '1.mp4')).exists(), false);
    expect(await File(path.join(directory.path, '2.mp4')).exists(), false);
    expect(repository.jobs, isEmpty);
    expect(repository.leases, ['media:start', 'media:end']);
    expect(directory.listSync().whereType<Directory>(), isEmpty);
    expect(merger.busy, false);
    expect(executor.probed, hasLength(2));
    expect(executor.probed.map(path.basename), ['1.mp4', '2.mp4']);
    expect(executor.commands.any((args) => args.contains('null')), isFalse);
    final conversions = executor.commands
        .where((args) => args.contains('libx264'))
        .toList();
    expect(conversions, hasLength(1));
    expect(
      conversions.every((args) => args.join(' ').contains('pad=360:640')),
      isTrue,
    );
    expect(statuses, contains('預備第 1 集'));
    expect(statuses, contains('第 1 集格式不符，轉檔中'));
    expect(statuses, contains('預備第 2 集'));
    expect(conversions.single, contains('ultrafast'));
    expect(executor.commands, hasLength(3));
    expect(statuses, contains('合併第 2 集中'));
  });

  test('media failure releases lease and preserves all source files', () async {
    final directory = await Directory.systemTemp.createTemp('guo-merge-test-');
    addTearDown(() => directory.delete(recursive: true));
    final repository = _MergeRepository(directory);
    final show = drama('a');
    repository.jobs = [downloadJob(show, 1), downloadJob(show, 2)];
    for (var number = 1; number <= 2; number++) {
      await File(path.join(directory.path, '$number.mp4'))
          .writeAsString('keep');
    }
    final merger = DownloadMerger(
      repository,
      executor: _MergeExecutor()..fail = true,
    );
    addTearDown(merger.dispose);
    await expectLater(
      merger.merge(repository.jobs!),
      throwsA(isA<AppFailure>()),
    );
    expect(
      await File(path.join(directory.path, '1.mp4')).readAsString(),
      'keep',
    );
    expect(directory.listSync().whereType<Directory>(), isEmpty);
    expect(repository.leases, ['media:start', 'media:end']);
    expect(merger.busy, false);
  });
  test('cleanup failure preserves the published merged output', () async {
    final directory = await Directory.systemTemp.createTemp(
      'guo-cleanup-test-',
    );
    addTearDown(() => directory.delete(recursive: true));
    final repository = _MergeRepository(directory)..failCleanup = true;
    final show = drama('a', title: '保存测试');
    repository.jobs = [downloadJob(show, 1), downloadJob(show, 2)];
    for (var number = 1; number <= 2; number++) {
      await File(path.join(directory.path, '$number.mp4'))
          .writeAsString('source');
    }
    final merger = DownloadMerger(repository, executor: _MergeExecutor());
    addTearDown(merger.dispose);
    await expectLater(
      merger.merge(repository.jobs!),
      throwsA(isA<AppFailure>()),
    );
    expect(
      await File(path.join(directory.path, '保存测试(1-2).mp4')).exists(),
      true,
    );
    expect(await File(path.join(directory.path, '1.mp4')).exists(), true);
    expect(repository.leases.last, 'media:end');
  });
  bool ffmpegAvailable;
  try {
    ffmpegAvailable = Process.runSync('ffmpeg', ['-version']).exitCode == 0;
  } catch (_) {
    ffmpegAvailable = false;
  }
  test(
    'real mixed-resolution merge is playable after deleting downloaded inputs',
    () async {
      final directory = await Directory.systemTemp.createTemp(
        'guo-real-merge-',
      );
      addTearDown(() => directory.delete(recursive: true));
      final repository = _MergeRepository(directory);
      final show = drama('synthetic', title: '本地合成');
      repository.jobs = [for (var n = 1; n <= 3; n++) downloadJob(show, n)];
      for (var n = 1; n <= 3; n++) {
        final result = await Process.run('ffmpeg', [
          '-v',
          'error',
          '-y',
          '-f',
          'lavfi',
          '-i',
          'testsrc2=size=${n == 3 ? '160x90' : '320x180'}:rate=12',
          '-f',
          'lavfi',
          '-i',
          'sine=frequency=440:sample_rate=44100',
          '-t',
          '1',
          '-c:v',
          'libx264',
          '-threads',
          '1',
          '-c:a',
          'aac',
          '-ac',
          '2',
          path.join(directory.path, '$n.mp4'),
        ]);
        expect(result.exitCode, 0, reason: '${result.stderr}');
      }
      final executor = _ProcessMergeExecutor();
      final merger = DownloadMerger(repository, executor: executor);
      addTearDown(merger.dispose);
      final output = await merger.merge(repository.jobs!);
      expect(path.basename(output), '本地合成(1-3).mp4');
      expect(executor.commands.any((args) => args.contains('null')), isFalse);
      final probe = await executor.probe(output);
      expect(probe.video['width'], 160);
      expect(probe.video['height'], 90);
      expect(repository.jobs, isEmpty);
      expect(directory.listSync().whereType<Directory>(), isEmpty);
      for (var n = 1; n <= 3; n++) {
        expect(
          await File(path.join(directory.path, '$n.mp4')).exists(),
          isFalse,
        );
      }
      final decoded = await Process.run('ffmpeg', [
        '-v',
        'error',
        '-xerror',
        '-i',
        output,
        '-f',
        'null',
        '-',
      ]);
      expect(decoded.exitCode, 0, reason: '${decoded.stderr}');
    },
    skip: ffmpegAvailable ? false : 'requires ffmpeg and ffprobe',
  );
}

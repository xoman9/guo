import 'dart:io';

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
  bool fail = false;
  @override
  Future<MediaProbe> probe(String file) async => MediaProbe({
    'format': {'duration': file.endsWith('merged.mp4') ? '24' : '12'},
    'streams': [
      {
        'codec_type': 'video',
        'codec_name': 'h264',
        'width': 720,
        'height': 1280,
      },
    ],
  });
  @override
  Future<void> run(
    List<String> arguments, {
    double duration = 0,
    void Function(double)? progress,
  }) async {
    if (fail) throw AppFailure('synthetic media failure');
    if (arguments.last != '-')
      await File(arguments.last).writeAsString('synthetic-video');
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
      continuousDownloadJobs([
        downloadJob(show, 2),
        downloadJob(show, 1),
      ]).first.episode.number,
      1,
    );
  });

  test('merge names are safe on Android and Windows', () {
    expect(downloadMergeName('  甲:剧/名?  '), '甲_剧_名_-合');
    expect(downloadMergeName('..'), '未命名-合');
    expect(downloadMergeName('CON'), '_CON-合');
  });

  test('merge publishes unique named output and retains originals', () async {
    final directory = await Directory.systemTemp.createTemp('guo-merge-test-');
    addTearDown(() => directory.delete(recursive: true));
    final repository = _MergeRepository(directory);
    final show = drama('a', title: '样本');
    repository.jobs = [downloadJob(show, 1), downloadJob(show, 2)];
    for (var number = 1; number <= 2; number++) {
      await File(
        path.join(directory.path, '$number.mp4'),
      ).writeAsString('source-$number');
    }
    final existing = File(path.join(directory.path, '样本-合.mp4'));
    await existing.writeAsString('keep-existing');
    final merger = DownloadMerger(repository, executor: _MergeExecutor());
    addTearDown(merger.dispose);
    final output = await merger.merge(repository.jobs!);
    expect(path.basename(output), '样本-合 (1).mp4');
    expect(await existing.readAsString(), 'keep-existing');
    expect(
      await File(path.join(directory.path, '1.mp4')).readAsString(),
      'source-1',
    );
    expect(repository.leases, ['media:start', 'media:end']);
    expect(directory.listSync().whereType<Directory>(), isEmpty);
    expect(merger.busy, false);
  });

  test('media failure releases lease and preserves all source files', () async {
    final directory = await Directory.systemTemp.createTemp('guo-merge-test-');
    addTearDown(() => directory.delete(recursive: true));
    final repository = _MergeRepository(directory);
    final show = drama('a');
    repository.jobs = [downloadJob(show, 1), downloadJob(show, 2)];
    await File(path.join(directory.path, '1.mp4')).writeAsString('keep');
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
}

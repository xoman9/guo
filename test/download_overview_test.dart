import 'package:duanju_app/download_overview.dart';
import 'package:duanju_app/local_store.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/fakes.dart';

DownloadOverview _watch(FakeRepository repository, LocalStore store) {
  final overview = DownloadOverview(
    repository,
    store,
    interval: const Duration(hours: 1),
  );
  addTearDown(overview.dispose);
  return overview;
}

void main() {
  final first = drama('a', title: '剧甲');
  final second = drama('b', title: '剧乙');

  test('download speed sums byte deltas over elapsed time', () {
    final sampler = DownloadSpeedSampler();
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 1000),
      downloadJob(second, 1, state: 'downloading', bytes: 2000),
    ], Duration.zero);
    expect(sampler.bytesPerSecond, 0);
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 3048),
      downloadJob(second, 1, state: 'downloading', bytes: 6096),
    ], const Duration(seconds: 2));
    expect(sampler.bytesPerSecond, 3072);
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 3048),
      downloadJob(second, 1, state: 'downloading', bytes: 6096),
    ], const Duration(seconds: 4));
    expect(sampler.bytesPerSecond, 0);
  });

  test('download speed ignores completed history and replaced counters', () {
    final sampler = DownloadSpeedSampler();
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 1000),
    ], Duration.zero);
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 3000),
      downloadJob(second, 1, bytes: 1 << 30),
    ], const Duration(seconds: 2));
    expect(sampler.bytesPerSecond, 1000);
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 10000000, created: 1),
    ], const Duration(seconds: 4));
    expect(sampler.bytesPerSecond, 0);
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 10, created: 1),
    ], const Duration(seconds: 6));
    expect(sampler.bytesPerSecond, 0);
  });

  test('pausing and reset clear the speed immediately', () {
    final sampler = DownloadSpeedSampler();
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 0),
    ], Duration.zero);
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 2048),
    ], const Duration(seconds: 2));
    expect(sampler.bytesPerSecond, 1024);
    sampler.update([
      downloadJob(first, 1, state: 'paused', bytes: 2048),
    ], const Duration(milliseconds: 2100));
    expect(sampler.bytesPerSecond, 0);
    sampler.reset();
    sampler.update([
      downloadJob(first, 1, state: 'downloading', bytes: 1 << 30),
    ], const Duration(seconds: 3));
    expect(sampler.bytesPerSecond, 0);
  });

  test('summarizes running and finished jobs', () async {
    final repository = FakeRepository()
      ..jobs = [
        downloadJob(first, 1, bytes: 300 << 20),
        downloadJob(first, 2, state: 'queued', progress: 0),
        downloadJob(second, 5, state: 'downloading', progress: .45),
      ];
    final overview = _watch(repository, await openStore());
    await pumpEventQueue();

    expect(overview.active, 2);
    expect(overview.current?.drama.id, second.id);
    expect(overview.current?.episode.number, 5);
    expect(overview.completed, 1);
    expect(overview.completedBytes, 300 << 20);
  });

  test('notifies only when the jobs change', () async {
    final repository = FakeRepository()..jobs = [downloadJob(first, 1)];
    final overview = _watch(repository, await openStore());
    var changes = 0;
    overview.addListener(() => changes++);
    await pumpEventQueue();
    await overview.refresh();
    expect(changes, 1);

    repository.jobs = [downloadJob(first, 1, state: 'queued', progress: 0)];
    await overview.refresh();
    expect(changes, 2);
    expect(overview.active, 1);
  });

  test('users without download permission are never polled', () async {
    final repository = FakeRepository()..jobs = [downloadJob(first, 1)];
    final overview = _watch(repository, await openStore(viewerPreferences));
    await pumpEventQueue();

    expect(repository.downloadRequests, 0);
    expect(overview.completed, 0);
  });
}

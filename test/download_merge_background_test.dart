import 'dart:async';

import 'package:duanju_app/core_bridge.dart';
import 'package:duanju_app/download_merge.dart';
import 'package:duanju_app/download_overview.dart';
import 'package:duanju_app/models.dart';
import 'package:duanju_app/ui/downloads/download_merge_progress.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/fakes.dart';

class _ControlledMerger extends DownloadMerger {
  _ControlledMerger(super.repository);
  final completion = Completer<String>();
  int cancellations = 0;
  bool ended = false;

  @override
  Future<String> merge(List<DownloadJob> selected) async {
    busy = true;
    this.drama = selected.first.drama;
    progress = .4;
    status = '合併第 1 集中';
    notifyListeners();
    try {
      output = await completion.future;
      status = '合併完成';
      return output!;
    } catch (failure) {
      error = failure.toString();
      status = '已取消合併';
      rethrow;
    } finally {
      busy = false;
      if (!ended) notifyListeners();
    }
  }

  @override
  void dispose() {
    ended = true;
    super.dispose();
  }

  @override
  Future<void> cancel() async {
    cancellations++;
    if (!completion.isCompleted) completion.completeError(AppFailure('已取消合併'));
  }
}

void main() {
  testWidgets(
    'closing progress and leaving the page keep the job alive; completion never pops another page',
    (tester) async {
      final repository = FakeRepository()..jobs = [downloadJob(drama('a'), 1)];
      final merger = _ControlledMerger(repository);
      final overview = DownloadOverview(
        repository,
        await openStore(),
        interval: const Duration(hours: 1),
        merger: merger,
      );
      addTearDown(overview.dispose);
      await tester.pump();
      overview.startMerge(repository.jobs!);
      expect(overview.mergeTask, same(merger));
      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) => Scaffold(
              body: Column(
                children: [
                  DownloadMergeBanner(merger: overview.mergeTask),
                  TextButton(
                    onPressed: () => Navigator.of(context).push(
                      MaterialPageRoute<void>(
                        builder: (_) => Scaffold(
                          appBar: AppBar(title: const Text('其他頁面')),
                          body: const Text('繼續瀏覽'),
                        ),
                      ),
                    ),
                    child: const Text('前往其他頁面'),
                  ),
                ],
              ),
            ),
          ),
        ),
      );
      await tester.tap(find.text('查看進度'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('close-merge-progress')));
      await tester.pumpAndSettle();
      expect(merger.busy, isTrue);
      expect(merger.cancellations, 0);
      expect(overview.merger, same(merger));
      await tester.tap(find.text('前往其他頁面'));
      await tester.pumpAndSettle();
      merger.completion.complete('/downloads/movie(1-1).mp4');
      await tester.pumpAndSettle();
      expect(find.text('繼續瀏覽'), findsOneWidget);
      expect(find.byType(AlertDialog), findsNothing);
      expect(merger.cancellations, 0);
      await tester.pageBack();
      await tester.pumpAndSettle();
      await tester.tap(find.text('查看進度'));
      await tester.pumpAndSettle();
      expect(find.text('合併完成'), findsWidgets);
      expect(find.text('/downloads/movie(1-1).mp4'), findsOneWidget);
      expect(find.text('複製路徑'), findsOneWidget);
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );

  testWidgets(
    'system back dismisses only the dialog; explicit cancel stops the job',
    (tester) async {
      final repository = FakeRepository()..jobs = [downloadJob(drama('a'), 1)];
      final merger = _ControlledMerger(repository);
      final overview = DownloadOverview(
        repository,
        await openStore(),
        interval: const Duration(hours: 1),
        merger: merger,
      );
      addTearDown(overview.dispose);
      await tester.pump();
      overview.startMerge(repository.jobs!);
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(body: DownloadMergeBanner(merger: overview.mergeTask)),
        ),
      );
      await tester.tap(find.text('查看進度'));
      await tester.pumpAndSettle();
      await tester.binding.handlePopRoute();
      await tester.pumpAndSettle();
      expect(find.byType(AlertDialog), findsNothing);
      expect(merger.busy, isTrue);
      expect(merger.cancellations, 0);
      await tester.tap(find.text('查看進度'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('取消合併'));
      await tester.pumpAndSettle();
      expect(merger.cancellations, 1);
      expect(merger.busy, isFalse);
      expect(merger.error, contains('已取消合併'));
      expect(find.text('合併未完成'), findsOneWidget);
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );

  test(
    'only the download manager owns the task and ends it on disposal',
    () async {
      final repository = FakeRepository()..jobs = [downloadJob(drama('a'), 1)];
      final merger = _ControlledMerger(repository);
      final overview = DownloadOverview(
        repository,
        await openStore(),
        interval: const Duration(hours: 1),
        merger: merger,
      );
      await pumpEventQueue();
      overview.startMerge(repository.jobs!);
      expect(
        () => overview.startMerge(repository.jobs!),
        throwsA(isA<AppFailure>()),
      );
      overview.dispose();
      await pumpEventQueue();
      expect(merger.cancellations, 1);
      expect(merger.busy, isFalse);
    },
  );
}

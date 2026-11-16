import 'dart:async';

import 'core_bridge.dart';
import 'models.dart';

/// Search all enabled, permitted sources without letting one broken endpoint
/// hide results from the others. A global search intentionally loads only the
/// first page from each source: unlike one-source search it must not fetch
/// another page from every provider during infinite scroll.
Future<CatalogPage> searchAcrossSources(
  AppRepository repository,
  Iterable<SourceSite> sources,
  String query, {
  Duration sourceTimeout = const Duration(seconds: 12),
  int concurrency = 4,
}) async {
  final targets = sources.where((site) => site.available).toList();
  if (query.trim().isEmpty || targets.isEmpty) {
    return CatalogPage(const []);
  }
  final collected = <Drama>[];
  final seen = <String>{};
  var succeeded = 0;
  var failed = 0;
  final limit = concurrency < 1 ? 1 : concurrency;
  for (var start = 0; start < targets.length; start += limit) {
    final group = targets.skip(start).take(limit).toList();
    final outcomes = await Future.wait(
      group.map((source) async {
        try {
          final result = await repository
              .catalog(source.id, query: query, page: 1)
              .timeout(sourceTimeout);
          return (source: source, page: result, failed: false);
        } catch (_) {
          return (source: source, page: CatalogPage(const <Drama>[]), failed: true);
        }
      }),
    );
    for (final outcome in outcomes) {
      if (outcome.failed) {
        failed++;
        continue;
      }
      succeeded++;
      for (final item in outcome.page.items) {
        // Do not merge results across different providers or lose the source
        // used by the detail page and the playback router.
        if (item.source == outcome.source.id && seen.add(item.id)) {
          collected.add(item);
        }
      }
    }
  }
  if (succeeded == 0 && failed > 0) {
    throw StateError('所有頻道搜尋失敗，請確認網路或稍後重試');
  }
  return CatalogPage(
    collected,
    hasMore: false,
    // Return successful channels even when others fail. A warning on
    // CatalogPage would turn DramaFeed into a retry loop, so the summary is
    // presented separately in the search UI in a future enhancement.
  );
}

import 'package:duanju_app/global_search.dart';
import 'package:duanju_app/models.dart';
import 'package:flutter_test/flutter_test.dart';
import 'support/fakes.dart';

class _PartiallyBrokenSearchRepository extends FakeRepository {
  @override
  Future<CatalogPage> catalog(
    String source, {
    int page = 1,
    String query = '',
    String category = '',
    bool force = false,
  }) async {
    if (source == 'broken') throw StateError('upstream unavailable');
    return super.catalog(
      source, page: page, query: query, category: category, force: force,
    );
  }
}

void main() {
  test('global search gathers results with original source IDs', () async {
    final repository = FakeRepository();
    final page = await searchAcrossSources(
      repository,
      const [
        SourceSite('hongguo', '红果', ''),
        SourceSite('dsd', '帝果', ''),
        SourceSite('disabled', '关闭', '', available: false),
      ],
      'test',
    );
    expect(page.hasMore, isFalse);
    expect(page.items.map((item) => item.source), ['hongguo', 'dsd']);
    expect(page.items.map((item) => item.id).toSet().length, 2);
    expect(repository.catalogRequests, [
      'hongguo||test|1',
      'dsd||test|1',
    ]);
  });

  test('one failed provider cannot hide results from others', () async {
    final repository = _PartiallyBrokenSearchRepository();
    final page = await searchAcrossSources(
      repository,
      const [
        SourceSite('broken', '失效', ''),
        SourceSite('dsd', '帝果', ''),
      ],
      'test',
    );
    expect(page.items.length, 1);
    expect(page.items.single.source, 'dsd');
  });

  test('all failed providers report an error', () async {
    final repository = _PartiallyBrokenSearchRepository();
    await expectLater(
      searchAcrossSources(
        repository,
        const [SourceSite('broken', '失效', '')],
        'test',
      ),
      throwsStateError,
    );
  });

  test('empty keyword does not contact providers', () async {
    final repository = FakeRepository();
    final page = await searchAcrossSources(
      repository,
      const [SourceSite('dsd', '帝果', '')],
      '  ',
    );
    expect(page.items, isEmpty);
    expect(repository.catalogRequests, isEmpty);
  });
}

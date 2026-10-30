import 'package:duanju_app/app_build.dart';
import 'package:duanju_app/models.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/fakes.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('guo adult bulk switch updates original source switches atomically', () async {
    expect(allSourcesEnabled, isTrue,
        reason: 'Run with --dart-define=ALL_SOURCES=true');
    final store = await openStore();
    final adult = SourceSite.values
        .where((site) => site.name.trimRight().endsWith('+'))
        .map((site) => site.id)
        .toList();
    expect(adult, contains('lujj31'));
    expect(adult, isNotEmpty);
    expect(store.sourceEnabled('hongguo'), isTrue);
    expect(adult.every((source) => !store.sourceEnabled(source)), isTrue);

    await store.setSourcesEnabled(adult, true);
    expect(adult.every(store.sourceEnabled), isTrue);
    expect(store.sourceEnabled('hongguo'), isTrue);
    expect(adult.every(restartStore(store).sourceEnabled), isTrue);

    await store.setSourceEnabled('lujj31', false);
    expect(store.sourceEnabled('lujj31'), isFalse);
    expect(store.sourceEnabled('huangdou'), isTrue);

    await store.setSourcesEnabled(adult, false);
    expect(adult.every((source) => !store.sourceEnabled(source)), isTrue);
    expect(store.sourceEnabled('hongguo'), isTrue);
    expect(adult.every((source) => !restartStore(store).sourceEnabled(source)),
        isTrue);
  });

  test('bulk toggle refuses unknown sources without changing configuration',
      () async {
    expect(allSourcesEnabled, isTrue,
        reason: 'Run with --dart-define=ALL_SOURCES=true');
    final store = await openStore();
    await expectLater(
      store.setSourcesEnabled(['lujj31', 'not-a-source'], true),
      throwsStateError,
    );
    expect(store.sourceEnabled('lujj31'), isFalse);
  });
}

import 'package:duanju_app/ui/app_scope.dart';
import 'package:duanju_app/ui/widgets/source_bar.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/fakes.dart';

void main() {
  testWidgets('source title opens chooser and old globe button is removed', (tester) async {
    var picked = 0;
    await tester.pumpWidget(
      MaterialApp(
        home: AppScope(
          repository: FakeRepository(),
          formFactor: FormFactor.phone,
          child: Scaffold(
            body: SourceBar(
              source: '露果 🔞',
              onSwitchSource: (_) => picked++,
              onOpenSearch: () {},
              onSearchSubmitted: (_) {},
            ),
          ),
        ),
      ),
    );
    expect(find.byIcon(Icons.public_rounded), findsNothing);
    expect(find.byIcon(Icons.keyboard_arrow_down_rounded), findsOneWidget);
    await tester.tap(find.text('露果 🔞'));
    await tester.pump();
    expect(picked, 1);
  });
}

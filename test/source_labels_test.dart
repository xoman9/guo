import 'package:duanju_app/models.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('specified existing source names end in the single adult symbol', () {
    for (final id in [
      'huangdou',
      'dsd',
      'huangju',
      'yeguo',
      'huangguo-video',
      'huangguoai',
      'cloudfront',
      'lujj31',
      'liuguo',
      'xiaoguo',
    ]) {
      final name = SourceSite.byId(id).name;
      expect(name.runes.last, 0x1f51e);
      expect(name.contains('+18'), false);
    }
    expect(SourceSite.hongguo.name, '红果');
    expect(SourceSite.byId('lujj31').id, 'lujj31');
    expect(SourceSite.isRestricted('lujj31'), isTrue);
    expect(SourceSite.isRestricted('liuguo'), isTrue);
    expect(SourceSite.isRestricted('xiaoguo'), isTrue);
    expect(SourceSite.values.any((site) => site.id == 'chaoguo'), isFalse);
    expect(SourceSite.values.any((site) => site.id == 'chengguo'), isFalse);
    // Chengguodj is a separate source and must not be removed by mistake.
    expect(SourceSite.values.any((site) => site.id == 'chengguodj'), isTrue);
  });
}

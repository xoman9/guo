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
    ]) {
      final name = SourceSite.byId(id).name;
      expect(name.runes.last, 0x1f51e);
      expect(name.contains('+18'), false);
    }
    expect(SourceSite.hongguo.name, '红果');
  });
}

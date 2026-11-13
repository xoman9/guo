import 'package:duanju_app/models.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('default-hidden channel name has exactly one trailing plus', () {
    final site = SourceSite.fromJson({
      'id': 'huangdou',
      'name': '黃豆',
      'description': '',
      'defaultOff': true,
      'available': true,
    });
    expect(site.id, 'huangdou');
    expect(site.defaultOff, isTrue);
    expect(site.name, '黃豆 +');
    expect(site.entryName, '黃豆 +');
  });

  test('previously plus-marked channel remains single plus', () {
    final site = SourceSite.fromJson({
      'id': 'lujj31',
      'name': '桃果 +',
      'description': '',
      'defaultOff': true,
      'available': true,
    });
    expect(site.name, '桃果 +');
  });

  test('normal original channel remains unmodified', () {
    final site = SourceSite.fromJson({
      'id': 'hongguo',
      'name': '紅果',
      'description': '',
      'defaultOff': false,
      'available': true,
    });
    expect(site.name, '紅果');
    expect(site.defaultOff, isFalse);
  });

  test('group entry of default-hidden channel also has suffix', () {
    final site = SourceSite.fromJson({
      'id': 'huangguoai',
      'name': '黃果 AI',
      'description': '',
      'group': 'huangguo',
      'groupName': '黃果',
      'entryName': 'AI',
      'defaultOff': true,
      'available': true,
    });
    expect(site.name, '黃果 AI +');
    expect(site.groupName, '黃果 +');
    expect(site.entryName, 'AI +');
  });
}

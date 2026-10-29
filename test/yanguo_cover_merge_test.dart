import 'package:duanju_app/models.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('Yanguo retains a working list poster when detail replaces metadata', () {
    const list = Drama(
      id: 'yanguo:601',
      source: 'yanguo',
      title: '原始片名',
      cover: 'https://xqxq1.cc/posters/601.jpg',
    );
    const playbackDetail = Drama(
      id: 'yanguo:601',
      source: 'yanguo',
      title: '正確影片名稱',
      cover: 'https://xqxq1.cc/placeholder.jpg',
    );
    final merged = list.merge(playbackDetail);
    expect(merged.cover, list.cover);
    expect(merged.title, playbackDetail.title);
  });

  test('Yanguo accepts a detail poster when list had none', () {
    const list = Drama(id: 'yanguo:602', source: 'yanguo', title: '片名');
    const detail = Drama(
      id: 'yanguo:602',
      source: 'yanguo',
      title: '片名',
      cover: 'https://xqxq1.cc/602.jpg',
    );
    expect(list.merge(detail).cover, detail.cover);
  });

  test('Other sources still update their posters normally', () {
    const list = Drama(
      id: 'luguo:603',
      source: 'luguo',
      title: '片名',
      cover: 'https://luyitian.com/old.jpg',
    );
    const detail = Drama(
      id: 'luguo:603',
      source: 'luguo',
      title: '片名',
      cover: 'https://luyitian.com/new.jpg',
    );
    expect(list.merge(detail).cover, detail.cover);
  });
}

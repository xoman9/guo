import 'dart:convert';

import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:duanju_app/models.dart';
import 'package:duanju_app/tvbox_js_source.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  const channel = MethodChannel('duanju/tvbox_js');

  setUp(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (call) async {
          if (call.method != 'execute') throw StateError('unexpected call');
          final args = Map<String, dynamic>.from(call.arguments as Map);
          final method = args['operation'];
          switch (method) {
            case 'homeContent':
              return jsonEncode({
                'class': [
                  {'type_id': 'new', 'type_name': '最近更新'},
                ],
                'list': [
                  {
                    'vod_id': 'AB-123',
                    'vod_name': 'Fixture title',
                    'vod_pic': 'https://images.example/cover.jpg',
                  },
                ],
              });
            case 'searchContent':
              return jsonEncode({
                'pagecount': 1,
                'list': [
                  {
                    'vod_id': 'XY-7',
                    'vod_name': 'Search title',
                    'vod_pic': 'https://images.example/search.jpg',
                  },
                ],
              });
            case 'detailContent':
              return jsonEncode({
                'list': [
                  {
                    'vod_id': 'AB-123',
                    'vod_name': 'Verified title',
                    'vod_pic': 'https://images.example/detail.jpg',
                    'vod_play_url':
                        '播放\$https://media.example/playlist.m3u8',
                  },
                ],
              });
            case 'playerContent':
              return jsonEncode({
                'url': 'https://media.example/master.m3u8',
                'headers': {
                  'Referer': 'https://embed.example/player',
                  'Origin': 'https://embed.example',
                  'User-Agent': 'Test WebView',
                },
              });
          }
          throw StateError('unexpected operation: $method');
        });
  });

  tearDown(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, null);
  });

  test('MissAV JS categories and catalog preserve source IDs', () async {
    final categories = await TVBoxJavaScriptSource.categories('missav');
    expect(categories.map((c) => c.id), ['', 'new']);
    final catalog = await TVBoxJavaScriptSource.catalog('missav');
    expect(catalog.items.single.id, 'missav:AB-123');
    expect(catalog.items.single.title, 'Fixture title');
  });

  test('JS search returns parsed cards, not raw TVBox JSON', () async {
    final page = await TVBoxJavaScriptSource.catalog(
      'tvbox-netflav',
      query: 'keyword',
    );
    expect(page.items.single.source, 'tvbox-netflav');
    expect(page.items.single.sourceId, 'XY-7');
    expect(page.hasMore, isFalse);
  });

  test('MissAV detail retains explicit direct HLS media', () async {
    const drama = Drama(
      id: 'missav:AB-123',
      source: 'missav',
      sourceId: 'AB-123',
      title: 'Catalog title',
    );
    final detail = await TVBoxJavaScriptSource.detail(drama);
    expect(detail.drama.title, 'Catalog title');
    expect(detail.episodes, hasLength(1));
    final media = await TVBoxJavaScriptSource.resolve(
      detail.drama,
      detail.episodes.single,
    );
    expect(media.url, 'https://media.example/playlist.m3u8');
  });

  test('Netflav JS playback only accepts explicit media URLs', () async {
    const drama = Drama(
      id: 'tvbox-netflav:XY-7',
      source: 'tvbox-netflav',
      sourceId: 'XY-7',
      title: 'Fixture title',
    );
    final episode = Episode({
      'id': 'tvbox-netflav:XY-7:1',
      'jsSourceIndex': 1,
    }, 1);
    final media = await TVBoxJavaScriptSource.resolve(drama, episode);
    expect(media.url, 'https://media.example/master.m3u8');
    expect(media.headers['Referer'], 'https://embed.example/player');
    expect(media.headers['Origin'], 'https://embed.example');
    expect(media.headers['User-Agent'], 'Test WebView');
  });
}

import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:flutter/services.dart';

import 'models.dart';

/// Adapter for the two bundled TVBox csp_GM JavaScript spiders.
/// The Android WebView executes the original script in its own first-party
/// page; no JavaScript evaluation or cookies are sent to the Go HTTP client.
class TVBoxJavaScriptSource {
  static const _channel = MethodChannel('duanju/tvbox_js');

  static bool supports(String source) =>
      Platform.isAndroid && (source == 'missav' || source == 'tvbox-netflav');

  // A single hidden Android WebView handles one request at a time.
  // Serialize adjacent category/search requests instead of reporting busy.
  static Future<void> _chain = Future<void>.value();

  static Future<Map<String, dynamic>> _execute(
    String source,
    String operation,
    List<Object?> args,
  ) {
    if (source == 'tvbox-netflav' && operation == 'playerContent') {
      // User-initiated playback takes priority over a slow detail request.
      // Cancel the currently running WebView before queuing playback; the
      // cancelled detail future will release _chain.
      // Capture the prior queue before updating _chain. Referring to
      // _chain inside this callback would await our own completion and
      // deadlock the request.
      final previous = _chain;
      final result = _channel.invokeMethod<void>('cancel').then(
        (_) => previous.then((_) => _perform(source, operation, args)),
      );
      _chain = result.then<void>((_) {}, onError: (Object _) {});
      return result;
    }
    final result = _chain.then((_) => _perform(source, operation, args));
    _chain = result.then<void>((_) {}, onError: (Object _) {});
    return result;
  }

  static Future<Map<String, dynamic>> _perform(
    String source,
    String operation,
    List<Object?> args,
  ) async {
    final raw = await _channel.invokeMethod<String>('execute', {
      'source': source,
      'operation': operation,
      'args': args,
    });
    if (raw == null || raw.trim().isEmpty) {
      throw StateError('$source JavaScript 未返回資料');
    }
    final decoded = jsonDecode(raw);
    if (decoded is! Map) {
      throw StateError('$source JavaScript 回應格式不正確');
    }
    final result = Map<String, dynamic>.from(decoded);
    if (result['error'] != null) {
      throw StateError('$source JavaScript：${result['error']}');
    }
    return result;
  }

  static List<Drama> _items(String source, Map<String, dynamic> response) {
    final rows = response['list'] as List? ?? const [];
    final seen = <String>{};
    return [
      for (final row in rows)
        if (row is Map)
          if (row['vod_tag'] != 'folder')
            if ((row['vod_id']?.toString() ?? '').isNotEmpty)
              if (seen.add(row['vod_id'].toString()))
                Drama(
                  id: '$source:${row['vod_id']}',
                  source: source,
                  sourceId: row['vod_id'].toString(),
                  title: row['vod_name']?.toString() ?? '',
                  cover: row['vod_pic']?.toString() ?? '',
                  description: row['vod_remarks']?.toString() ?? '',
                  episodes: 1,
                ),
    ];
  }

  static Future<List<CatalogCategory>> categories(String source) async {
    final result = await _execute(source, 'homeContent', [false]);
    return [
      CatalogCategory.all,
      for (final row in result['class'] as List? ?? const [])
        if (row is Map)
          CatalogCategory(
            row['type_id']?.toString() ?? '',
            row['type_name']?.toString() ?? '',
          ),
    ];
  }

  static Future<CatalogPage> catalog(
    String source, {
    int page = 1,
    String category = '',
    String query = '',
  }) async {
    Map<String, dynamic> result;
    if (query.trim().isNotEmpty) {
      result = await _execute(source, 'searchContent', [query.trim(), false, page]);
    } else if (page == 1 && category.isEmpty) {
      result = await _execute(source, 'homeContent', [false]);
    } else {
      final tid = category.isNotEmpty
          ? category
          : source == 'missav'
              ? 'new'
              : 'trending?';
      result = await _execute(source, 'categoryContent', [tid, page, false, <String, Object>{}]);
    }
    final pageCount = intValue(result['pagecount']);
    final items = _items(source, result);
    return CatalogPage(
      items,
      hasMore: pageCount > 0 ? page < pageCount : false,
      page: page,
      fresh: true,
    );
  }

  static Future<DramaDetail> detail(Drama drama) async {
    final response = await _execute(
      drama.source,
      'detailContent',
      [[drama.sourceId]],
    );
    final value = response['list'] is List && (response['list'] as List).isNotEmpty
        ? (response['list'] as List).first
        : response;
    if (value is! Map) throw StateError('JavaScript 未返回影片詳情');
    final data = Map<String, dynamic>.from(value);
    final updated = drama.merge(Drama(
      id: drama.id,
      source: drama.source,
      sourceId: drama.sourceId,
      // MissAV's legacy TVBox spider returns an uppercased ID as vod_name.
      // Prefer the verified catalog title instead of replacing it with a
      // one-character/placeholder name from the detail scraper.
      title: drama.source == 'missav' &&
              drama.title.trim().length > 1 &&
              drama.title.trim() != '['
          ? drama.title
          : (data['vod_name']?.toString().trim().isNotEmpty == true
              ? data['vod_name'].toString().trim()
              : drama.title),
      cover: data['vod_pic']?.toString() ?? drama.cover,
      description: data['vod_content']?.toString() ?? '',
      episodes: 1,
    ));
    final chapters = <Episode>[];
    final playURL = data['vod_play_url']?.toString() ?? '';
    if (playURL.isNotEmpty) {
      for (final item in playURL.split('#')) {
        final parts = item.split(r'$');
        final url = parts.length >= 2 ? parts.last : '';
        if (url.startsWith('https://') || url.startsWith('http://')) {
          chapters.add(Episode({
            'id': '${drama.id}:${chapters.length + 1}',
            'source': drama.source,
            'title': parts.first,
            'currentEpisode': chapters.length + 1,
            'jsMediaURL': url,
          }, chapters.length + 1));
        }
      }
    }
    final playData = data['vod_play_data'] as List? ?? const [];
    for (final entry in playData) {
      if (entry is! Map) continue;
      final sources = entry['media'] as List? ?? const [];
      for (final media in sources) {
        if (media is! Map) continue;
        final ext = media['ext'];
        final replacement = ext is Map ? ext['replace'] : null;
        final index = replacement is Map ? intValue(replacement['src']) : chapters.length + 1;
        chapters.add(Episode({
          'id': '${drama.id}:${chapters.length + 1}',
          'source': drama.source,
          'title': entry['from']?.toString() ?? '播放',
          'currentEpisode': chapters.length + 1,
          'jsSourceIndex': index,
        }, chapters.length + 1));
      }
    }
    // The present Netflav site often returns an empty srcs list: it is
    // functioning as a metadata/promotional catalog. Do not manufacture a
    // playable first episode when the site returned no playable sources.
    if (chapters.isEmpty && drama.source == 'tvbox-netflav') {
      return DramaDetail(updated, chapters);
    }
    if (chapters.isEmpty) {
      chapters.add(Episode({
        'id': '${drama.id}:1',
        'source': drama.source,
        'title': '播放',
        'currentEpisode': 1,
        'jsSourceIndex': 1,
      }, 1));
    }
    return DramaDetail(updated, chapters);
  }

  static Future<PlaybackPlan> resolve(Drama drama, Episode episode) async {
    final referer = drama.source == 'missav'
        ? 'https://missav.ws/cn/${Uri.encodeComponent(drama.sourceId)}'
        : 'https://netflav.com/video?id=${Uri.encodeQueryComponent(drama.sourceId)}';
    final headers = <String, String>{'Referer': referer};
    final direct = episode.raw['jsMediaURL']?.toString() ?? '';
    if (direct.startsWith('https://') || direct.startsWith('http://')) {
      return PlaybackPlan(url: direct, headers: headers);
    }
    // The csp_GM player can request an iframe and return {type:match}.
    // The platform adapter must observe a real HTTPS media request, never
    // invent a URL or treat the player page as a downloadable stream.
    final result = await _execute(
      drama.source,
      'playerContent',
      [drama.sourceId, intValue(episode.raw['jsSourceIndex'])],
    );
    final url = result['url']?.toString() ?? '';
    if (!url.startsWith('https://') && !url.startsWith('http://')) {
      throw StateError('JS 播放器尚未取得可播放的媒體網址');
    }
    // Preserve media request headers observed by the Android WebView;
    // cross-origin iframe players can require their own Referer/Origin.
    final observed = result['headers'];
    if (observed is Map) {
      for (final entry in observed.entries) {
        final name = entry.key.toString();
        if (name == 'Referer' || name == 'Origin' || name == 'User-Agent') {
          final value = entry.value.toString();
          if (value.isNotEmpty) headers[name] = value;
        }
      }
    }
    return PlaybackPlan(url: url, headers: headers);
  }
}

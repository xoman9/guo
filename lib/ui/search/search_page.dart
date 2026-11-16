import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../core_bridge.dart';
import '../../drama_feed.dart';
import '../../global_search.dart';
import '../../local_store.dart';
import '../../models.dart';
import '../../source_catalog.dart';
import '../app_scope.dart';
import '../detail/detail_page.dart';
import '../shell/app_shell.dart';
import '../theme/motion.dart';
import '../widgets/drama_grid.dart';
import '../widgets/follow_button.dart';
import '../widgets/focus_ring.dart';
import '../widgets/motion/motion_widgets.dart';
import '../widgets/search_field.dart';
import '../widgets/status_view.dart';

class SearchPage extends StatefulWidget {
  const SearchPage({
    super.key,
    required this.store,
    required this.source,
    this.query = '',
  });
  final LocalStore store;
  final SourceSite source;
  final String query;

  @override
  State<SearchPage> createState() => _SearchPageState();
}

class _SearchPageState extends State<SearchPage> {
  late final _controller = TextEditingController(text: widget.query);
  final _focus = FocusNode();
  late AppRepository _repository;
  late SourceCatalog _catalog;
  DramaFeed? _results;
  String _submitted = '';
  List<String> _suggestions = const [];
  Timer? _debounce;
  int _suggestion = 0;
  final _shownSuggestions = <Object>{};
  bool _ready = false;
  bool _global = true;
  String _sourceFilter = '';

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_ready) return;
    _ready = true;
    _repository = AppScope.of(context).repository;
    _catalog = SourceCatalog(_repository, widget.source);
    if (widget.query.trim().isNotEmpty) _submit(widget.query);
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _suggestion++;
    unawaited(_repository.cancelSuggestions());
    _results?.dispose();
    _controller.dispose();
    _focus.dispose();
    super.dispose();
  }

  void _forget(String query) {
    unawaited(HapticFeedback.selectionClick());
    unawaited(widget.store.forgetSearch(query).catchError((Object _) {}));
  }

  void _submit(String text) {
    final query = text.trim();
    if (query.isEmpty) return;
    _controller.text = query;
    _focus.unfocus();
    _debounce?.cancel();
    _suggestion++;
    unawaited(widget.store.rememberSearch(query).catchError((Object _) {}));
    _results?.dispose();
    final results = _results = DramaFeed(
      load: (page, _) => _global
          ? searchAcrossSources(_repository, widget.store.sources, query)
          : _catalog.search(query, page: page),
    )..addListener(() => setState(() {}));
    unawaited(results.reload());
    setState(() {
      _submitted = query;
      _sourceFilter = '';
      _suggestions = const [];
    });
  }

  void _changed(String text) {
    _debounce?.cancel();
    final query = text.trim();
    final generation = ++_suggestion;
    if (_global || !widget.source.searchSuggestions || query.isEmpty) {
      setState(() => _suggestions = const []);
      return;
    }
    _debounce = Timer(const Duration(milliseconds: 250), () async {
      try {
        final items = await _repository.suggestions(query);
        if (mounted && generation == _suggestion) {
          setState(() => _suggestions = items);
        }
      } catch (_) {}
    });
  }

  void _open(Drama drama) => openPage<void>(
    context,
    (_) => DetailPage(store: widget.store, drama: drama),
  );

  @override
  Widget build(BuildContext context) {
    final field = SearchField(
      hint: _global ? '搜尋所有已開啟頻道' : '在 ${widget.source.name} 中搜索',
      controller: _controller,
      focusNode: _focus,
      autofocus: widget.query.isEmpty,
      onChanged: _changed,
      onSubmitted: _submit,
      letterKeyboard: widget.store.televisionKeyboard,
      matches: !_global && widget.source.searchSuggestions ? _repository.suggestions : null,
    );
    if (AppScope.of(context).television) {
      return Scaffold(
        body: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(
                televisionInset,
                27,
                televisionInset,
                0,
              ),
              child: field,
            ),
            _modeSelector(),
            Expanded(child: FadeSwitcher(child: _body(context))),
          ],
        ),
      );
    }
    return Scaffold(
      appBar: AppBar(
        titleSpacing: 0,
        title: Padding(padding: const EdgeInsets.only(right: 16), child: field),
      ),
      body: SafeArea(
        top: false,
        bottom: false,
        child: Column(children: [
          _modeSelector(),
          Expanded(child: FadeSwitcher(child: _body(context))),
        ]),
      ),
    );
  }

  Widget _modeSelector() => Padding(
    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
    child: Row(children: [
      ChoiceChip(
        label: const Text('全頻道搜尋'),
        selected: _global,
        onSelected: (_) => _changeMode(true),
      ),
      const SizedBox(width: 8),
      ChoiceChip(
        label: Text(widget.source.name),
        selected: !_global,
        onSelected: (_) => _changeMode(false),
      ),
    ]),
  );

  void _changeMode(bool global) {
    if (_global == global) return;
    _results?.dispose();
    _results = null;
    _suggestion++;
    _debounce?.cancel();
    setState(() {
      _global = global;
      _suggestions = const [];
      _sourceFilter = '';
    });
    if (_submitted.isNotEmpty) _submit(_submitted);
  }

  Widget _body(BuildContext context) {
    final television = AppScope.of(context).television;
    if (_suggestions.isNotEmpty && _controller.text.trim() != _submitted) {
      return ListView(
        key: const ValueKey('suggest'),
        padding: television
            ? const EdgeInsets.fromLTRB(televisionInset, 8, televisionInset, 27)
            : const EdgeInsets.symmetric(vertical: 8),
        children: [
          for (final (index, item) in _suggestions.indexed)
            EnterOnce(
              key: ValueKey(index),
              id: item,
              played: _shownSuggestions,
              index: index,
              child: TileFocus(
                builder: (focusNode) => ListTile(
                  focusNode: focusNode,
                  leading: const Icon(Icons.search_rounded),
                  title: Text(item),
                  onTap: () => _submit(item),
                ),
              ),
            ),
        ],
      );
    }
    final results = _results;
    if (results == null) {
      return KeyedSubtree(
        key: const ValueKey('recent'),
        child: _recent(context),
      );
    }
    if (results.items.isEmpty) {
      if (results.loading) {
        return const StatusView.loading(
          key: ValueKey('loading'),
          title: '正在搜索',
        );
      }
      if (results.error.isNotEmpty) {
        return StatusView(
          key: const ValueKey('error'),
          title: '搜索失败',
          message: results.error,
          icon: Icons.cloud_off_rounded,
          action: '重试',
          onAction: () => results.reload(),
        );
      }
      return StatusView(
        key: const ValueKey('empty'),
        title: '没有找到「$_submitted」',
        message: '换个关键词试试',
        icon: Icons.search_off_rounded,
      );
    }
    final inset = television
        ? televisionInset
        : AppScope.formFactorOf(context) == FormFactor.phone
        ? 16.0
        : 24.0;
    final channels = _global
        ? <String, int>{
            for (final source in widget.store.sources)
              if (results.items.any((item) => item.source == source.id))
                source.id: results.items.where((item) => item.source == source.id).length,
          }
        : const <String, int>{};
    final visibleItems = _sourceFilter.isEmpty
        ? results.items
        : results.items.where((item) => item.source == _sourceFilter).toList();
    return Column(children: [
      if (_global && channels.isNotEmpty)
        SizedBox(height: 48, child: ListView(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          scrollDirection: Axis.horizontal,
          children: [
            _sourceChip('', '全部 (${results.items.length})'),
            for (final entry in channels.entries)
              _sourceChip(entry.key, '${SourceSite.byId(entry.key).name} (${entry.value})'),
          ],
        )),
      Expanded(child: LayoutBuilder(
      key: ValueKey(('results', _submitted, _global, _sourceFilter)),
      builder: (context, constraints) => AutoLoadMore(
        feed: results,
        child: CustomScrollView(
          slivers: [
            SliverPadding(
              padding: EdgeInsets.fromLTRB(
                inset,
                television ? 16 : 8,
                inset,
                0,
              ),
              sliver: DramaSliverGrid(
                width: constraints.maxWidth - inset * 2,
                items: visibleItems,
                autofocus: television,
                heroScope: 'drama',
                onOpen: _open,
                onMore: (drama) => quickFollow(context, widget.store, drama),
              ),
            ),
            SliverPadding(
              padding: EdgeInsets.fromLTRB(
                inset,
                8,
                inset,
                television ? 27 : 16,
              ),
              sliver: SliverToBoxAdapter(child: FeedFooter(feed: results)),
            ),
          ],
        ),
      ),
      ),
      ),
    ]);
  }

  Widget _sourceChip(String id, String label) => Padding(
    padding: const EdgeInsets.only(right: 8),
    child: ChoiceChip(
      label: Text(label),
      selected: _sourceFilter == id,
      onSelected: (_) => setState(() => _sourceFilter = id),
    ),
  );

  Widget _recent(BuildContext context) => ListenableBuilder(
    listenable: widget.store,
    builder: (context, _) {
      final recent = widget.store.recentSearches;
      return FadeSwitcher(child: _recentContent(context, recent));
    },
  );

  Widget _recentContent(BuildContext context, List<String> recent) {
    if (recent.isEmpty) {
      return StatusView(
        key: const ValueKey('empty'),
        title: _global ? '搜尋所有已開啟頻道' : '在 ${widget.source.name} 中搜索',
        message: '输入剧名或关键词',
        icon: Icons.search_rounded,
      );
    }
    final theme = Theme.of(context);
    return ListView(
      key: const ValueKey('recent'),
      padding: AppScope.of(context).television
          ? const EdgeInsets.fromLTRB(televisionInset, 8, televisionInset, 27)
          : const EdgeInsets.fromLTRB(16, 8, 16, 24),
      children: [
        Row(
          children: [
            Expanded(
              child: Text(
                '最近搜索',
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
            FocusRing(
              radius: 999,
              child: TextButton(
                onPressed: () => widget.store.clearRecentSearches().catchError(
                  (Object _) {},
                ),
                child: const Text('清空'),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        AnimatedSize(
          duration: Motion.of(context, Motion.medium),
          curve: Motion.enter,
          alignment: Alignment.topLeft,
          child: Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final query in recent)
                FocusRing(
                  radius: 8,
                  child: GestureDetector(
                    onLongPress: () => _forget(query),
                    onSecondaryTap: () => _forget(query),
                    child: ActionChip(
                      avatar: const Icon(Icons.history_rounded, size: 16),
                      label: Text(query),
                      onPressed: () => _submit(query),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }
}

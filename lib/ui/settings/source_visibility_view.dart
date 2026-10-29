import 'package:flutter/material.dart';

import '../../local_store.dart';
import '../../models.dart';
import '../widgets/focus_ring.dart';
import 'settings_rows.dart';

/// oktv-style visibility switch, independent from source permissions.
class SourceVisibilityView extends StatefulWidget {
  const SourceVisibilityView({super.key, required this.store});
  final LocalStore store;

  @override
  State<SourceVisibilityView> createState() => _SourceVisibilityViewState();
}

class _SourceVisibilityViewState extends State<SourceVisibilityView> {
  final Set<String> _pending = {};
  bool _adultPending = false;

  Future<void> _setVisible(String source, bool visible) async {
    if (_pending.contains(source)) return;
    setState(() => _pending.add(source));
    try {
      await saveSetting(context,
          () => widget.store.setSiteVisible(source, visible));
    } finally {
      if (mounted) setState(() => _pending.remove(source));
    }
  }

  Future<void> _setAdultVisible(List<SourceSite> sources, bool visible) async {
    if (_adultPending) return;
    setState(() => _adultPending = true);
    try {
      await saveSetting(context, () => widget.store.setSitesVisible(
        sources.map((site) => site.id), visible));
    } finally {
      if (mounted) setState(() => _adultPending = false);
    }
  }

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: widget.store,
    builder: (context, _) {
      // Visibility remains configurable for permitted but disabled sources.
      final sources = widget.store.permittedSources;
      final adult = sources.where((site) => site.name.contains('🔞')).toList();
      final adultHidden =
          adult.every((site) => !widget.store.siteVisible(site.id));
      final visibleCount =
          sources.where((site) => widget.store.siteVisible(site.id)).length;
      return ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  '源顯示列表 · $visibleCount/${sources.length}',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              ),
              FilledButton(
                key: const ValueKey('visibility-adult-toggle'),
                onPressed: _adultPending || adult.isEmpty
                    ? null
                    : () => _setAdultVisible(adult, adultHidden),
                child: Text(adultHidden ? '成人頻道-開' : '成人頻道-關'),
              ),
            ],
          ),
          const SizedBox(height: 12),
          const Text(
            '右側開關只控制首頁及站源選擇列表的顯示，不影響站源權限、已下載檔案或觀看紀錄。'
            '尚未啟用的站源仍需在「站源啟用管理」開啟才可使用。',
          ),
          const SizedBox(height: 16),
          for (final group in SourceGroup.fromSources(sources))
            Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: SettingsCard(
                children: [
                  if (group.sources.length > 1)
                    ListTile(
                      title: Text(group.name),
                      leading: const Icon(Icons.hub_outlined),
                    ),
                  for (final site in group.sources)
                    TileFocus(
                      builder: (focusNode) => SwitchListTile(
                        key: ValueKey('visibility-${site.id}'),
                        focusNode: focusNode,
                        contentPadding:
                            const EdgeInsets.fromLTRB(16, 0, 16, 0),
                        secondary: const Icon(Icons.dns_outlined),
                        title: Text(site.entryName),
                        subtitle: Text(site.description),
                        value: widget.store.siteVisible(site.id),
                        onChanged: _pending.contains(site.id)
                            ? null
                            : (value) => _setVisible(site.id, value),
                      ),
                    ),
                ],
              ),
            ),
        ],
      );
    },
  );
}

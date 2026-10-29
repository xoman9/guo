import 'package:flutter/material.dart';

import '../../local_store.dart';
import '../../models.dart';
import '../app_scope.dart';
import '../widgets/focus_ring.dart';
import 'settings_rows.dart';

class SourcesView extends StatefulWidget {
  const SourcesView({super.key, required this.store});
  final LocalStore store;

  @override
  State<SourcesView> createState() => _SourcesViewState();
}

class _SourcesViewState extends State<SourcesView> {
  bool _bulkChanging = false;

  Future<void> _toggleAdultSources(
    BuildContext context,
    List<SourceSite> adult,
    bool enable,
  ) async {
    if (_bulkChanging) return;
    final repository = AppScope.of(context).repository;
    final wasEnabled = [
      for (final site in adult)
        if (widget.store.sourceEnabled(site.id)) site.id,
    ];
    setState(() => _bulkChanging = true);
    try {
      await saveSetting(context, () async {
        await widget.store.setSourcesEnabled(
          adult.map((site) => site.id),
          enable,
        );
        if (!enable && repository.supportsDownloads) {
          for (final id in wasEnabled) {
            await repository.controlDownloads('pauseSource', id: id);
          }
        }
      });
    } finally {
      if (mounted) setState(() => _bulkChanging = false);
    }
  }

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: widget.store,
    builder: (context, _) {
      final store = widget.store;
      final admin = store.profile.admin;
      final repository = AppScope.of(context).repository;
      final adult = [
        for (final site in SourceSite.values)
          if (site.name.contains('🔞')) site,
      ];
      final adultOff = adult.every(
        (site) => !store.sourceEnabled(site.id),
      );
      return ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 4, 4, 12),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    '站源列表',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ),
                FilledButton(
                  key: const ValueKey('source-adult-bulk-toggle'),
                  style: FilledButton.styleFrom(
                    backgroundColor: const Color(0xFFFFDCD6),
                    foregroundColor: const Color(0xFF692D25),
                    disabledBackgroundColor: const Color(0xFFF5DFDA),
                    shape: const StadiumBorder(),
                  ),
                  onPressed: !admin || _bulkChanging || adult.isEmpty
                      ? null
                      : () => _toggleAdultSources(context, adult, adultOff),
                  child: Text(adultOff ? '成人頻道-開' : '成人頻道-關'),
                ),
              ],
            ),
          ),
          if (!admin) ...[
            const _Note('需要管理员才能开启或关闭站源'),
            const SizedBox(height: 12),
          ],
          SettingsCard(
            children: [
              for (final site in SourceSite.values)
                _SourceTile(
                  site: site,
                  enabled: store.sourceEnabled(site.id),
                  onChanged: admin
                      ? (value) => saveSetting(context, () async {
                          await store.setSourceEnabled(site.id, value);
                          if (!value && repository.supportsDownloads) {
                            await repository.controlDownloads(
                              'pauseSource',
                              id: site.id,
                            );
                          }
                        })
                      : null,
                ),
            ],
          ),
          const SizedBox(height: 12),
          const _Note(
            '右上方按钮可一次开启或关闭所有带 🔞 标记的站源。关闭的站源不在发现、搜索、追剧、观看记录和下载中显示，记录保留，重新开启后恢复。正在下载的集会暂停，重新开启后可在下载页继续。',
          ),
        ],
      );
    },
  );
}

class _SourceTile extends StatelessWidget {
  const _SourceTile({
    required this.site,
    required this.enabled,
    required this.onChanged,
  });
  final SourceSite site;
  final bool enabled;
  final ValueChanged<bool>? onChanged;

  @override
  Widget build(BuildContext context) => MergeSemantics(
    child: TileFocus(
      builder: (focusNode) => ListTile(
        focusNode: focusNode,
        minTileHeight: 68,
        contentPadding: const EdgeInsets.fromLTRB(20, 4, 16, 4),
        title: Text(site.name),
        subtitle: Text(site.description),
        onTap: onChanged == null ? null : () => onChanged!(!enabled),
        trailing: ExcludeFocus(
          child: Switch(value: enabled, onChanged: onChanged),
        ),
      ),
    ),
  );
}

class _Note extends StatelessWidget {
  const _Note(this.text);
  final String text;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(horizontal: 16),
    child: Text(
      text,
      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
        color: Theme.of(context).colorScheme.onSurfaceVariant,
        height: 1.6,
      ),
    ),
  );
}

import 'package:flutter/material.dart';

import '../../app_device.dart';
import '../../local_store.dart';
import '../../models.dart';
import '../app_scope.dart';
import '../widgets/focus_ring.dart';
import 'settings_rows.dart';
import 'sorani_account_view.dart';

class SourcesView extends StatelessWidget {
  const SourcesView({super.key, required this.store});
  final LocalStore store;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: store,
    builder: (context, _) {
      final admin = store.profile.admin;
      final repository = AppScope.of(context).repository;
      return ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
        children: [
          if (!admin) ...[
            const _Note('需要管理员才能开启或关闭站源'),
            const SizedBox(height: 12),
          ],
          SettingsCard(
            children: [
              SwitchListTile(
                title: const Text('隱藏頻道'),
                subtitle: Text(
                  '預設關閉；開啟時一次啟用所有 + 頻道（目前已啟用 '
                  '${SourceSite.values.where((site) => site.defaultOff && store.sourceEnabled(site.id)).length}'
                  '／${SourceSite.values.where((site) => site.defaultOff).length}）。開啟時全部啟用，關閉時全部停用。',
                ),
                value: store.hideAdditionalSources,
                onChanged: admin
                    ? (value) => saveSetting(
                        context,
                        () => store.setHideAdditionalSources(value),
                      )
                    : null,
              ),
              const Divider(height: 1),
              for (final site in SourceSite.values) ...[
                if (site.id == 'missav' || site.id == 'tvbox-jable')
                  ListTile(
                    title: Text('${site.displayName} · 驗證網站'),
                    subtitle: const Text('HTTP 403 時，使用 Android WebView 手動完成站方驗證'),
                    leading: const Icon(Icons.verified_user_outlined),
                    enabled: admin,
                    onTap: admin ? () async {
                      await saveSetting(context, () async {
                        final session = await AppDevice.verifySiteInBrowser(site.id);
                        if (session == null) return;
                        await repository.saveSiteVerification(
                          site.id, session.cookie, session.userAgent,
                        );
                        if (!context.mounted) return;
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(content: Text(
                            '網站驗證資訊已儲存。請重新整理頻道；若仍顯示 403，站點可能不允許原生 HTTP 連線。',
                          )),
                        );
                      });
                    } : null,
                  ),
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
            ],
          ),
          const SizedBox(height: 12),
          const _Note(
            '关闭的站源不在发现、搜索、追剧、观看记录和下载中显示，记录保留，重新开启后恢复。正在下载的集会暂停，重新开启后可在下载页继续。',
          ),
        ],
      );
    },
  );
}

class SourceAccountsView extends StatelessWidget {
  const SourceAccountsView({super.key, required this.store});
  final LocalStore store;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: store,
    builder: (context, _) {
      final admin = store.profile.admin;
      return ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
        children: [
          if (!admin) ...[
            const _Note('需要管理员才能登录或退出站源账号'),
            const SizedBox(height: 12),
          ],
          SettingsCard(children: [SoraniAccountTile(admin: admin)]),
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
        title: Text(site.displayName),
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

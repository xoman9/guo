import 'dart:io';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:path_provider/path_provider.dart';

import '../../local_store.dart';
import '../app_scope.dart';
import '../widgets/focus_ring.dart';
import '../widgets/group_card.dart';
import '../widgets/status_view.dart';
import 'settings_rows.dart';

String storageSize(int bytes) {
  if (bytes < 0) return '暂不可用';
  if (bytes < 1 << 20) return '${(bytes / 1024).toStringAsFixed(1)} KB';
  if (bytes < 1 << 30) return '${(bytes / (1 << 20)).toStringAsFixed(1)} MB';
  return '${(bytes / (1 << 30)).toStringAsFixed(2)} GB';
}

class StorageView extends StatefulWidget {
  const StorageView({super.key, required this.store});
  final LocalStore store;

  @override
  State<StorageView> createState() => _StorageViewState();
}

class _StorageViewState extends State<StorageView> {
  Map<String, dynamic>? _info;
  String? _error;
  bool _busy = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_info == null && _error == null) _refresh();
  }

  Future<void> _refresh() async {
    try {
      final info = await AppScope.of(context).repository.storage();
      if (mounted) {
        setState(() {
          _info = info;
          _error = null;
        });
      }
    } catch (failure) {
      if (mounted) setState(() => _error = '$failure');
    }
  }

  Future<String?> _pickParent() async {
    if (!Platform.isAndroid && !Platform.isIOS) {
      return FilePicker.getDirectoryPath(dialogTitle: '选择下载保存位置');
    }
    final support = await getApplicationSupportDirectory();
    final directories = <String, String>{support.path: '应用内部存储'};
    if (Platform.isAndroid) {
      final external = await getExternalStorageDirectories() ?? [];
      for (final (index, directory) in external.indexed) {
        directories[directory.path] = index == 0
            ? '设备共享存储（应用目录）'
            : 'SD 卡 ${index + 1}（应用目录）';
      }
    } else {
      final documents = await getApplicationDocumentsDirectory();
      directories[documents.path] = '文件 App 可见目录';
    }
    if (!mounted) return null;
    return showOptionDialog(
      context,
      title: '选择下载位置',
      options: directories,
      selected: null,
    );
  }

  Future<bool> _authorizeSharedStorage() async {
    if (!Platform.isAndroid) return true;
    const channel = MethodChannel('duanju/device');
    final granted =
        await channel.invokeMethod<bool>('sharedStorageAccess') ?? false;
    if (granted) return true;
    await channel.invokeMethod<bool>('requestSharedStorageAccess');
    if (mounted) setState(() => _error = '请到系统设置允许所有文件存取，返回后重新选择资料夹');
    return false;
  }

  Future<void> _change({
    bool sharedDefault = false,
    bool appPrivate = false,
  }) async {
    if (_busy) return;
    try {
      String? selected;
      if (appPrivate) {
        selected = '${(await getApplicationSupportDirectory()).path}/downloads';
      } else if (Platform.isAndroid) {
        if (!await _authorizeSharedStorage()) return;
        const channel = MethodChannel('duanju/device');
        selected = await channel.invokeMethod<String>(
          sharedDefault ? 'sharedDownloadDirectory' : 'selectDownloadDirectory',
        );
      } else {
        selected = await _pickParent();
        if (Platform.isIOS && selected != null)
          selected = '$selected/guo-downloads';
      }
      if (selected == null || !mounted) return;
      if (sharedDefault || appPrivate || Platform.isIOS)
        await Directory(selected).create(recursive: true);
      if (!mounted) return;
      final confirmed = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('更换下载资料夹'),
          content: Text(
            '日后下载使用：\n$selected\n\n不搬移或删除原有文件。下载列表只显示所选目录的记录，切回原目录可恢复查看。请先暂停全部下载。',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('取消'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('使用此资料夹'),
            ),
          ],
        ),
      );
      if (confirmed != true || !mounted) return;
      setState(() {
        _busy = true;
        _error = null;
      });
      final repository = AppScope.of(context).repository;
      await repository.changeDownloadDirectory(selected);
      final active = await repository.downloadDirectory();
      if (mounted) setState(() => _info = {...?_info, 'directory': active});
      await _refresh();
    } catch (failure) {
      if (mounted) setState(() => _error = '$failure');
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _move() async {
    if (_busy) return;
    final parent = await _pickParent();
    if (parent == null || !mounted) return;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('迁移已下载内容？'),
        content: Text(
          '将下载目录里的全部内容迁移到：\n$parent\n\n下载会先暂停，复制成功后清理旧目录。请保证目标有足够空间，并在完成后继续下载。',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          FilledButton(
            autofocus: true,
            onPressed: () => Navigator.pop(context, true),
            child: const Text('迁移'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    final repository = AppScope.of(context).repository;
    final navigator = Navigator.of(context, rootNavigator: true);
    showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (context) => const PopScope(
        canPop: false,
        child: AlertDialog(
          content: Row(
            spacing: 20,
            children: [
              CircularProgressIndicator(),
              Expanded(child: Text('正在迁移，请保持应用运行…')),
            ],
          ),
        ),
      ),
    );
    setState(() => _busy = true);
    try {
      await repository.moveDownloads(parent);
    } catch (failure) {
      if (mounted) setState(() => _error = '$failure');
    } finally {
      navigator.pop();
      if (mounted) setState(() => _busy = false);
    }
    await _refresh();
  }

  @override
  Widget build(BuildContext context) {
    final info = _info;
    if (info == null) {
      return _error == null
          ? const StatusView.loading()
          : StatusView(
              title: '无法读取下载目录',
              message: _error,
              icon: Icons.folder_off_outlined,
              action: '重试',
              onAction: _refresh,
            );
    }
    final text = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;
    final directory = info['directory'] as String? ?? '';
    final admin = widget.store.profile.admin;
    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
      children: [
        GroupCard(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                spacing: 4,
                children: [
                  Text(
                    '已使用 ${storageSize((info['bytes'] as num?)?.toInt() ?? 0)}',
                    style: text.headlineSmall,
                  ),
                  Text(
                    '剩余 ${storageSize((info['free'] as num?)?.toInt() ?? -1)} · ${info['files'] ?? 0} 个文件',
                    style: text.bodyMedium?.copyWith(
                      color: colors.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
            const Divider(indent: 16, endIndent: 16),
            SettingsTile(
              title: '当前下载目录',
              subtitle: directory,
              trailing: FocusRing(
                radius: 999,
                child: IconButton(
                  tooltip: '复制路径',
                  onPressed: () =>
                      Clipboard.setData(ClipboardData(text: directory)),
                  icon: const Icon(Icons.copy_rounded),
                ),
              ),
            ),
            const Divider(indent: 16, endIndent: 16),
            SettingsTile(
              title: '更换下载资料夹',
              subtitle: admin ? '使用系统选择器；保留原目录，可切回查看' : '需要管理员',
              enabled: admin && !_busy,
              onTap: () => _change(),
            ),
            if (Platform.isAndroid) ...[
              const Divider(indent: 16, endIndent: 16),
              SettingsTile(
                title: '使用共享下载目录',
                subtitle: 'Download/guo',
                enabled: admin && !_busy,
                onTap: () => _change(sharedDefault: true),
              ),
              const Divider(indent: 16, endIndent: 16),
              SettingsTile(
                title: '授权所有文件存取',
                enabled: admin && !_busy,
                onTap: () async {
                  try {
                    await _authorizeSharedStorage();
                  } catch (failure) {
                    if (mounted) setState(() => _error = '$failure');
                  }
                },
              ),
            ],
            const Divider(indent: 16, endIndent: 16),
            SettingsTile(
              title: '使用原应用下载目录',
              subtitle: '切回应用内部 downloads，保留共享资料夹的文件',
              enabled: admin && !_busy,
              onTap: () => _change(appPrivate: true),
            ),
            const Divider(indent: 16, endIndent: 16),
            SettingsTile(
              title: '更改并迁移目录',
              subtitle: admin ? '下载会先暂停，复制成功后清理旧目录' : '需要管理员',
              enabled: admin && !_busy,
              onTap: _move,
            ),
            const Divider(indent: 16, endIndent: 16),
            SettingsTile(title: '刷新用量', enabled: !_busy, onTap: _refresh),
          ],
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
          child: Text(
            _error ??
                (_busy
                    ? '正在更新下载资料夹…'
                    : '原始文件保存在 剧名/0001/ 等分集目录；转换后的 第1集.mp4 位于剧名目录，成功转换后会删除该集原始影片与金钥，失败则保留以便重试。更换目录不搬移旧文件，原有迁移功能仍可使用。'),
            style: text.bodyMedium?.copyWith(
              color: _error == null ? colors.onSurfaceVariant : colors.error,
            ),
          ),
        ),
      ],
    );
  }
}

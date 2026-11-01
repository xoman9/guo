import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../download_merge.dart';

Future<void> showDownloadMergeProgress(
  BuildContext context,
  DownloadMerger merger,
) => showDialog<void>(
  context: context,
  builder: (context) => ListenableBuilder(
    listenable: merger,
    builder: (context, _) => AlertDialog(
      title: Text(
        merger.busy
            ? '合併進度'
            : merger.error != null
            ? '合併未完成'
            : '合併完成',
      ),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (merger.drama != null) Text(merger.drama!.title),
            const SizedBox(height: 12),
            if (merger.busy) LinearProgressIndicator(value: merger.progress),
            const SizedBox(height: 12),
            Text(merger.status),
            if (merger.error != null) ...[
              const SizedBox(height: 8),
              Text(merger.error!),
            ],
            if (merger.output != null) ...[
              const SizedBox(height: 8),
              SelectableText(merger.output!),
            ],
            if (merger.busy) ...[
              const SizedBox(height: 12),
              const Text('關閉視窗或返回其他頁面會繼續合併。成品儲存成功後才清理原分集；可回下載頁查看進度。'),
            ],
          ],
        ),
      ),
      actions: [
        if (merger.busy)
          TextButton(
            onPressed: merger.cancellable ? merger.cancel : null,
            child: const Text('取消合併'),
          ),
        if (merger.output != null)
          TextButton(
            onPressed: () =>
                Clipboard.setData(ClipboardData(text: merger.output!)),
            child: const Text('複製路徑'),
          ),
        FilledButton(
          key: const ValueKey('close-merge-progress'),
          onPressed: () => Navigator.of(context).pop(),
          child: Text(merger.busy ? '關閉視窗，繼續合併' : '關閉'),
        ),
      ],
    ),
  ),
);

class DownloadMergeBanner extends StatelessWidget {
  const DownloadMergeBanner({super.key, required this.merger});
  final DownloadMerger? merger;

  @override
  Widget build(BuildContext context) {
    final task = merger;
    if (task == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: task,
      builder: (context, _) {
        if (task.drama == null) return const SizedBox.shrink();
        return Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Card(
            child: Column(
              children: [
                ListTile(
                  leading: Icon(
                    task.busy
                        ? Icons.merge_rounded
                        : task.error != null
                        ? Icons.error_outline
                        : Icons.task_alt_rounded,
                  ),
                  title: Text(
                    task.drama!.title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  subtitle: Text(
                    task.status,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  trailing: TextButton(
                    onPressed: () => showDownloadMergeProgress(context, task),
                    child: const Text('查看進度'),
                  ),
                ),
                if (task.busy) LinearProgressIndicator(value: task.progress),
                if (!task.busy)
                  Align(
                    alignment: Alignment.centerRight,
                    child: TextButton(
                      onPressed: task.clearResult,
                      child: const Text('關閉提示'),
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }
}

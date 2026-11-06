import 'dart:io';

import 'package:ffmpeg_kit_flutter_new_min_gpl/ffmpeg_kit_config.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:path/path.dart' as path;
import 'package:path_provider/path_provider.dart';

import '../../app_device.dart';
import '../../core_bridge.dart';
import '../../download_merge.dart';
import '../../models.dart';
import '../app_scope.dart';
import '../widgets/focus_ring.dart';

Future<void> mergeDownloads(
  BuildContext context,
  List<DownloadJob> jobs,
) async {
  final messenger = ScaffoldMessenger.of(context);
  final merge = DownloadMerge(AppScope.of(context).repository);
  // The merge job is not owned by the dialog. Closing its UI must not cancel
  // conversion or dispose of the running FFmpeg executor.
  final operation = _run(merge, jobs).then((result) => result).catchError(
    (Object error) => merge.cancelled ? '已取消合并' : '$error',
  );
  operation.then((message) {
    if (message != null && messenger.mounted) {
      messenger.showSnackBar(SnackBar(content: Text(message)));
    }
  }).whenComplete(merge.dispose);
  await showDialog<void>(
    context: context,
    builder: (context) => _MergeDialog(merge, operation),
  );
}

Future<String?> _run(DownloadMerge merge, List<DownloadJob> jobs) async {
  await merge.prepare(jobs);
  if (Platform.isIOS) {
    final directory = Directory(
      path.join((await getApplicationDocumentsDirectory()).path, '合并视频'),
    );
    await directory.create(recursive: true);
    final file = _unusedFile(directory.path, merge.fileName);
    await merge.write(file, file: file);
    return '已保存到「文件」App 里本应用的「合并视频」文件夹';
  }
  if (Platform.isAndroid) {
    final uri = await AppDevice.createDocument(merge.fileName, 'video/mp4');
    if (uri == null) return null;
    try {
      final output = await FFmpegKitConfig.getSafParameterForWrite(uri);
      if (output == null) throw AppFailure('无法写入所选位置');
      await merge.write(output);
    } catch (_) {
      await AppDevice.deleteDocument(uri);
      rethrow;
    }
    return '合并完成';
  }
  final saved = await FilePicker.saveFile(
    fileName: merge.fileName,
    bytes: Uint8List(0),
    mimeType: 'video/mp4',
    dialogTitle: '保存合并视频',
  );
  if (saved == null) return null;
  final file = saved.toFilePath();
  await merge.write(file, file: file);
  return '已保存到 $file';
}

String _unusedFile(String directory, String name) {
  final stem = path.basenameWithoutExtension(name);
  var file = path.join(directory, name);
  for (var attempt = 2; File(file).existsSync(); attempt++) {
    file = path.join(directory, '$stem ($attempt).mp4');
  }
  return file;
}

class _MergeDialog extends StatefulWidget {
  const _MergeDialog(this.merge, this.operation);
  final DownloadMerge merge;
  final Future<String?> operation;

  @override
  State<_MergeDialog> createState() => _MergeDialogState();
}

class _MergeDialogState extends State<_MergeDialog> {
  @override
  void initState() {
    super.initState();
    widget.operation.then((_) {
      if (mounted) Navigator.pop(context);
    });
  }

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: widget.merge,
    builder: (context, _) => AlertDialog(
      title: const Text('合并视频'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: 16,
        children: [
          Text(widget.merge.status),
          LinearProgressIndicator(value: widget.merge.progress),
          const Text('可以关闭此窗口，合并仍会继续执行。'),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('背景继续'),
        ),
        FocusRing(
          radius: 20,
          child: TextButton(
            autofocus: true,
            onPressed: widget.merge.cancelled
                ? null
                : () {
                    widget.merge.cancel();
                    Navigator.pop(context);
                  },
            child: const Text('取消合并'),
          ),
        ),
      ],
    ),
  );
}

import 'dart:io';

import 'package:flutter/services.dart';

import 'media_pipeline.dart';
import 'models.dart';

Future<AvcHardwareEncoder?> findAndroidAvcEncoder(MergePlan plan) async {
  if (!Platform.isAndroid ||
      plan.video.videoCodec != 'h264' ||
      !plan.videoChanges.any((change) => change))
    return null;
  final video = plan.video.video;
  if (!{'yuv420p', 'nv12'}.contains(video['pix_fmt']) ||
      {'smpte2084', 'arib-std-b67'}.contains(video['color_transfer']))
    return null;
  final width = intValue(video['width']), height = intValue(video['height']);
  final rate = '${video['r_frame_rate'] ?? '30/1'}'.split('/');
  var fps =
      (double.tryParse(rate.first) ?? 30) /
      (rate.length == 2 ? double.tryParse(rate.last) ?? 1 : 1);
  if (!fps.isFinite || fps <= 0) fps = 30;
  final bitrate = (width * height * fps * .2).round().clamp(500000, 12000000);
  try {
    final result = await const MethodChannel('duanju/device')
        .invokeMapMethod<String, dynamic>('h264HardwareEncoder', {
          'width': width,
          'height': height,
          'frameRate': fps,
          'bitrate': bitrate,
        });
    if (result == null ||
        result['name'] is! String ||
        !{'nv12', 'yuv420p'}.contains(result['pixelFormat']))
      return null;
    return AvcHardwareEncoder(
      name: result['name'] as String,
      bitrate: intValue(result['bitrate']),
      bitrateMode: intValue(result['bitrateMode']),
      pixelFormat: result['pixelFormat'] as String,
    );
  } on PlatformException {
    return null;
  } on MissingPluginException {
    return null;
  }
}

package com.duanju.duanju_app

import android.media.MediaCodecInfo
import android.media.MediaCodecList
import android.os.Build

object HardwareVideoEncoder {
    fun findAvc(width: Int, height: Int, frameRate: Double, bitrate: Int): Map<String, Any>? {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.Q) return null
        if (width <= 0 || height <= 0 || !frameRate.isFinite() || frameRate <= 0) return null
        val codecs = runCatching { MediaCodecList(MediaCodecList.REGULAR_CODECS).codecInfos }.getOrNull() ?: return null
        for (codec in codecs) {
            if (!codec.isEncoder || codec.supportedTypes.none { it.equals("video/avc", ignoreCase = true) }) continue
            val hardware = codec.isHardwareAccelerated && !codec.isSoftwareOnly
            if (!hardware) continue
            val selected = runCatching {
                val capabilities = codec.getCapabilitiesForType("video/avc")
                val video = capabilities.videoCapabilities ?: return@runCatching null
                if (!video.areSizeAndRateSupported(width, height, frameRate)) return@runCatching null
                val pixel = when {
                    capabilities.colorFormats.contains(MediaCodecInfo.CodecCapabilities.COLOR_FormatYUV420SemiPlanar) -> "nv12"
                    capabilities.colorFormats.contains(MediaCodecInfo.CodecCapabilities.COLOR_FormatYUV420Planar) -> "yuv420p"
                    else -> return@runCatching null
                }
                val encoder = capabilities.encoderCapabilities ?: return@runCatching null
                val mode = when {
                    encoder.isBitrateModeSupported(MediaCodecInfo.EncoderCapabilities.BITRATE_MODE_VBR) -> MediaCodecInfo.EncoderCapabilities.BITRATE_MODE_VBR
                    encoder.isBitrateModeSupported(MediaCodecInfo.EncoderCapabilities.BITRATE_MODE_CBR) -> MediaCodecInfo.EncoderCapabilities.BITRATE_MODE_CBR
                    else -> return@runCatching null
                }
                mapOf("name" to codec.name, "pixelFormat" to pixel, "bitrateMode" to mode,
                    "bitrate" to bitrate.coerceIn(video.bitrateRange.lower, video.bitrateRange.upper))
            }.getOrNull()
            if (selected != null) return selected
        }
        return null
    }
}

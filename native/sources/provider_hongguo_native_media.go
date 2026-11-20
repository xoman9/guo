package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"duanjuapp/native/core"
)

func (p *hongguoProvider) resolveHongguoAppMedia(ctx context.Context, videoID string) (core.ProviderMedia, error) {
	if !hongguoNumericID.MatchString(videoID) {
		return core.ProviderMedia{}, errors.New("红果视频 ID 无效")
	}
	// Prefer the current video-tab API. The older video_model/v1 route can
	// return HTTP 200 with an empty payload even for valid episodes.
	tabQuery := url.Values{
		"video_id": {videoID}, "video_id_type": {"0"},
		"device_level": {"1"}, "need_all_video_definition": {"1"},
		"need_mp4_align": {"0"}, "use_os_player": {"0"},
		"video_platform": {"1024"}, "source": {"0"}, "from_video_id": {""},
	}
	tabResult, tabErr := p.hongguoAppRequest(ctx, http.MethodGet, "/reading/bookapi/video_tab/multi_video_model/v1/", tabQuery, nil)
	if tabErr == nil {
		data := nestedMap(tabResult, "data", "video_model_data", videoID)
		if len(data) == 0 {
			tabErr = errors.New("红果 video_tab 未返回所请求分集")
		} else {
			media, err := selectHongguoAppMedia(hongguoDecodeVideoModel(data))
			if err == nil { return media, nil }
			tabErr = err
		}
	}
	if err := ctx.Err(); err != nil { return core.ProviderMedia{}, err }
	// Updated player interface, not the retired video_model/v1 endpoint.
	payload := map[string]any{
		"biz_param": map[string]any{
			"detail_page_version": 0, "device_level": 1,
			"disable_digg_stat": false, "disable_video_relate_book": false,
			"need_all_video_definition": true, "need_mp4_align": false,
			"use_os_player": false, "use_server_dns": false, "video_platform": 1024,
		},
		"mixed_video_id_map": map[string]any{"1004": []string{videoID}},
		"dr_scene": "player",
	}
	playerResult, playerErr := p.hongguoAppRequest(ctx, http.MethodPost, "/novel/player/multi_video_model/player/v1", nil, payload)
	if playerErr == nil {
		entries := nestedMap(playerResult, "data")
		if group, ok := entries["1004"].(map[string]any); ok { entries = group }
		data, _ := entries[videoID].(map[string]any)
		if len(data) == 0 {
			playerErr = errors.New("红果播放器接口未返回所请求分集")
		} else {
			media, err := selectHongguoAppMedia(hongguoDecodeVideoModel(data))
			if err == nil { return media, nil }
			playerErr = err
		}
	}
	return core.ProviderMedia{}, fmt.Errorf("video_tab: %v；播放器接口: %v", tabErr, playerErr)
}

func hongguoDecodeVideoModel(data map[string]any) map[string]any {
	model, _ := data["video_model"].(map[string]any)
	if encoded, ok := data["video_model"].(string); ok && encoded != "" {
		if decoded, err := decodeHongguoBase64(encoded); err == nil && len(decoded) > 0 {
			encoded = string(decoded)
		}
		var parsed map[string]any
		decoder := json.NewDecoder(strings.NewReader(encoded))
		decoder.UseNumber()
		if decoder.Decode(&parsed) == nil { model = parsed }
	}
	if len(model) == 0 && len(anyList(data["video_list"])) > 0 {
		model = map[string]any{"video_list": data["video_list"], "video_duration": data["video_duration"], "duration": data["duration"]}
	}
	return model
}

func selectHongguoAppMedia(model map[string]any) (core.ProviderMedia, error) {
	variants := anyList(model["video_list"])
	if rows, ok := model["video_list"].(map[string]any); ok && len(variants) == 0 {
		keys := make([]string, 0, len(rows))
		for key := range rows {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			variants = append(variants, rows[key])
		}
	}
	duration, _ := strconv.ParseFloat(core.MapString(model, "video_duration", "duration"), 64)
	var keyErr error
	type scoredMedia struct {
		media core.ProviderMedia
		score int
	}
	var choices []scoredMedia
	for _, row := range variants {
		variant, _ := row.(map[string]any)
		meta := nestedMap(variant, "video_meta")
		codec := strings.ToLower(core.MapString(meta, "codec_type"))
		if codec == "bytevc2" || strings.Contains(strings.ToLower(core.MapString(variant, "gear_des_key")), "bytevc2") {
			continue
		}
		addresses := hongguoMediaAddresses(variant)
		if len(addresses) == 0 {
			continue
		}
		media := core.ProviderMedia{Referer: "https://novel.snssdk.com/", Duration: time.Duration(duration * float64(time.Second))}
		encryption := nestedMap(variant, "encrypt_info")
		spade := core.MapString(encryption, "spade_a")
		if spade != "" || encryption["encrypt"] == true || core.MapString(encryption, "encryption_method") == "cenc-aes-ctr" {
			var err error
			media.CENCKey, err = hongguoContentKey(spade)
			if err != nil {
				keyErr = err
				continue
			}
		}
		height, _ := strconv.Atoi(core.MapString(meta, "vheight"))
		if definition, err := strconv.Atoi(hongguoQualityNumber.FindString(core.MapString(meta, "definition"))); err == nil && definition > 0 {
			height = definition
		} else if width, _ := strconv.Atoi(core.MapString(meta, "vwidth")); width > 0 && (height == 0 || width < height) {
			height = width
		}
		media.Quality = height
		quality := height * 10
		if codec == "h264" || codec == "avc1" {
			quality++
		}
		for _, address := range addresses {
			media.URL = address
			choices = append(choices, scoredMedia{media: media, score: quality})
		}
	}
	if len(choices) > 0 {
		sort.SliceStable(choices, func(i, j int) bool { return choices[i].score > choices[j].score })
		selected := choices[0].media
		for _, choice := range choices {
			selected.Variants = append(selected.Variants, choice.media)
		}
		return selected, nil
	}
	if keyErr != nil {
		return core.ProviderMedia{}, fmt.Errorf("红果 App 媒体密钥不可用: %w", keyErr)
	}
	return core.ProviderMedia{}, errors.New("红果 App 未返回兼容的媒体，已跳过不支持的编码")
}

func hongguoMediaAddresses(info map[string]any) []string {
	var addresses []string
	seen := map[string]bool{}
	var add func(any)
	add = func(value any) {
		switch value := value.(type) {
		case string:
			address := strings.TrimSpace(value)
			if len(address) > 8192 {
				return
			}
			if !core.IsProviderHTTPMediaURL(address) {
				decoded, err := decodeHongguoBase64(address)
				if err != nil {
					return
				}
				address = strings.TrimSpace(string(decoded))
			}
			if core.IsProviderHTTPMediaURL(address) && !seen[address] {
				seen[address] = true
				addresses = append(addresses, address)
			}
		case []any:
			for _, item := range value {
				add(item)
			}
		}
	}
	for _, key := range []string{"main_url", "backup_url", "backup_url_1", "backup_url_2", "backup_urls", "url_list"} {
		add(info[key])
	}
	return addresses
}

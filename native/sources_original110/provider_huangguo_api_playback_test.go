package sources

import "testing"

func TestHuangguoAIPlaybackPageMediaURL(t *testing.T) {
    const page = "https://huangguoai.com/video/abc123/"
    cases := []struct {
        name, html, want string
    }{
        {"hls", `<video data-hls="https://cdn.example.com/film/master.m3u8?token=abc&amp;ep=1"></video>`, "https://cdn.example.com/film/master.m3u8?token=abc&ep=1"},
        {"mp4", `<video data-hls="https://cdn.example.com/film.mp4"></video>`, "https://cdn.example.com/film.mp4"},
        {"missing", `<iframe src="https://ads.example.com"></iframe>`, ""},
        {"javascript", `<video data-hls="javascript:alert(1)"></video>`, ""},
        {"http", `<video data-hls="http://cdn.example.com/film.m3u8"></video>`, ""},
        {"html-not-media", `<video data-hls="https://ads.example.com/watch"></video>`, ""},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            if got := huangguoAIPageMediaURL(page, tc.html); got != tc.want {
                t.Fatalf("got %q; want %q", got, tc.want)
            }
        })
    }
}

func TestHuangguoAIPlayMediaURLAliases(t *testing.T) {
    const base = "https://huangguoai.com/"
    cases := []struct {
        name string
        data map[string]any
        want string
    }{
        {"original", map[string]any{"video_url": "https://cdn.example/one.m3u8"}, "https://cdn.example/one.m3u8"},
        {"camelCase", map[string]any{"videoUrl": "https://cdn.example/two.mp4"}, "https://cdn.example/two.mp4"},
        {"nested", map[string]any{"data": map[string]any{"play_url": "https://cdn.example/three.m3u8"}}, "https://cdn.example/three.m3u8"},
        {"relative", map[string]any{"url": "/videos/four.mp4"}, "https://huangguoai.com/videos/four.mp4"},
        {"adHtml", map[string]any{"url": "https://ads.example/watch"}, ""},
        {"http", map[string]any{"video_url": "http://cdn.example/five.m3u8"}, ""},
        {"script", map[string]any{"url": "javascript:alert(1)"}, ""},
        {"none", map[string]any{}, ""},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            if got := huangguoAIPlayMediaURL(base, tc.data); got != tc.want {
                t.Fatalf("got %q; want %q", got, tc.want)
            }
        })
    }
}

func TestHuangguoAIDetailFallbackOnlyEpisodeOne(t *testing.T) {
    const base = "https://huangguoai.com/"
    detail := map[string]any{"video_url": "https://yd-lphu-hls.hzjvlo.cn/sample/index.m3u8?auth_key=example"}
    want := "https://yd-lphu-hls.hzjvlo.cn/sample/index.m3u8?auth_key=example"
    if got := huangguoAIDetailFirstEpisodeMediaURL(base, 1, detail); got != want {
        t.Fatalf("first episode SSR URL = %q; want %q", got, want)
    }
    if got := huangguoAIDetailFirstEpisodeMediaURL(base, 2, detail); got != "" {
        t.Fatalf("episode two incorrectly reused first episode URL: %q", got)
    }
    if got := huangguoAIDetailFirstEpisodeMediaURL(base, 1, map[string]any{"video_url": "https://ads.example.com/watch"}); got != "" {
        t.Fatalf("invalid URL was accepted: %q", got)
    }
}

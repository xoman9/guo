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

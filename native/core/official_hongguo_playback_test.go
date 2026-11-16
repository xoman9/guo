package core

import (
    "context"
    "strings"
    "testing"
)

type offlineHongguoJSHost struct{ Host }
func (*offlineHongguoJSHost) FinishMedia(_ context.Context, media ProviderMedia, source string) (ProviderMedia, error) {
    if source != "hongguo" {
        panic("unexpected source "+source)
    }
    return media, nil
}

func TestOfficialHongguoV102ModulePlaybackDispatchOffline(t *testing.T) {
    mod := officialModules["hongguo"]
    if mod == nil || mod.manifest.Version != "1.0.2" {
        t.Fatalf("Hongguo official JS 1.0.2 not loaded")
    }
    d := &Downloader{}
    p, ok := d.provider("hongguo")
    if !ok {
        t.Fatal("hongguo provider not registered")
    }
    if js, isJS := p.(*jsProvider); !isJS || js.module != mod {
        t.Fatal("hongguo still using legacy Go playback provider")
    }
    host := &offlineHongguoJSHost{}
    p = mod.provider(host)
    // An already resolved public URL is forwarded through the same
    // host.finishMedia integration used by the upstream module.
    input := Task{
        DramaID: "hongguo:123",
        Chapter: Chapter{Source: "hongguo", VideoURL: "https://cdn.example.com/episode1.mp4", Referer: "https://hongguoduanju.com/"},
    }
    media, err := p.ResolveMedia(context.Background(), input)
    if err != nil {
        t.Fatalf("JS direct playback bridge failed: %v", err)
    }
    if media.URL != input.Chapter.VideoURL || media.Referer != input.Chapter.Referer {
        t.Fatalf("JS media bridge damaged playback URL or referer: %+v", media)
    }
    // Invalid CENC episode ids must fail before any remote request. Never
    // substitute a preview or a different episode.
    input.Chapter.VideoURL = "hongguo-cenc://invalid"
    _, err = p.ResolveMedia(context.Background(), input)
    if err == nil || !strings.Contains(err.Error(), "ID") {
        t.Fatalf("invalid CENC episode should be rejected, got %v", err)
    }
}

package core

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestLujj31SourceRegistration(t *testing.T) {
	if !isHuangguoProviderSource(sourceLujj31) || canonicalProviderSource("lujj31.lol") != sourceLujj31 {
		t.Fatal("lujj31 must have a distinct registered source identity")
	}
	for _, address := range []string{"https://lujj31.lol/", "https://www.lujj31.lol/"} {
		if providerSourceForURL(address) != sourceLujj31 {
			t.Fatalf("unknown host: %s", address)
		}
	}
	address := "https://lujj31.lol/index.php/vod/type/id/1.html"
	candidates := (&Downloader{cfg: defaultConfig()}).providerURLCandidates(address)
	if len(candidates) != 1 || candidates[0] != address {
		t.Fatalf("unexpected rewriting: %v", candidates)
	}
}

func TestLujj31OktvHTMLRoutes(t *testing.T) {
	var paths []string
	d := sourceFixtureDownloader(t, func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		if strings.Contains(r.URL.Path, "/api.php/provide/") {
			t.Fatalf("old JSON API used: %s", r.URL)
		}
		var body string
		switch r.URL.Path {
		case "/":
			body = `<html><body><nav><a href="/index.php/vod/type/id/1.html">最新</a><a href="/index.php/vod/type/id/2.html">影片</a></nav></body></html>`
		case "/index.php/vod/type/id/1.html":
			body = `<html><body><div class="module-item"><a href="/index.php/vod/detail/id/123.html" title="合成影片"><img data-src="/poster.jpg"></a></div></body></html>`
		case "/index.php/vod/type/id/1/page/2.html":
			body = `<html><body><div class="module-item"><a href="/index.php/vod/detail/id/124.html" title="第二部">第二部</a></div></body></html>`
		case "/index.php/vod/detail/id/123.html":
			body = `<html><body><h1 class="page-title">合成影片</h1><div class="module-play-list-content"><a href="/index.php/vod/play/id/123/sid/1/nid/1.html">第1集</a><a href="/index.php/vod/play/id/123/sid/1/nid/2.html">第2集</a><a href="/index.php/vod/play/id/123/sid/2/nid/1.html">备用1</a><a href="/index.php/vod/play/id/999/sid/1/nid/1.html">其他影片</a></div></body></html>`
		default:
			return nil, fmt.Errorf("unexpected request: %s", r.URL.Path)
		}
		return sourceFixtureResponse(r, http.StatusOK, body), nil
	})
	categories, err := d.fetchLujj31Categories(context.Background())
	if err != nil || len(categories) != 2 || categories[0].ID != "1" {
		t.Fatalf("categories: %+v %v", categories, err)
	}
	items, more, err := d.fetchLujj31CatalogPage(context.Background(), 1, "", "")
	if err != nil || !more || len(items) != 1 || items[0].ID != "lujj31:123" {
		t.Fatalf("catalog: %+v %t %v", items, more, err)
	}
	items, more, err = d.fetchLujj31CatalogPage(context.Background(), 2, "", "")
	if err != nil || !more || len(items) != 1 || items[0].ID != "lujj31:124" {
		t.Fatalf("second page: %+v %t %v", items, more, err)
	}
	drama, chapters, err := d.fetchLujj31Detail(context.Background(), "123")
	if err != nil || drama.ID != "lujj31:123" || len(chapters) != 2 || chapters[1].PageURL != "https://lujj31.lol/index.php/vod/play/id/123/sid/1/nid/2.html" {
		t.Fatalf("detail: %+v %+v %v", drama, chapters, err)
	}
	for _, p := range paths {
		if strings.Contains(p, "/api.php/") {
			t.Fatal("wrong API path used", paths)
		}
	}
}

func TestLujj31PlayerAndIdentity(t *testing.T) {
	direct := `<script>var player_aaaa={"id":123,"sid":1,"nid":2,"encrypt":0,"url":"https://cdn.test/ep2.m3u8"};</script>`
	url, err := lujj31PlayerURL(direct, "123")
	if err != nil || url != "https://cdn.test/ep2.m3u8" {
		t.Fatalf("player: %s %v", url, err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte("https://cdn.test/ep3.mp4"))
	url, err = lujj31PlayerURL(`player_data={"id":123,"encrypt":2,"url":"`+encoded+`"};`, "123")
	if err != nil || url != "https://cdn.test/ep3.mp4" {
		t.Fatalf("encrypted player: %s %v", url, err)
	}
	if _, err := lujj31PlayerURL(`player_data={"id":123,"encrypt":0,"url":"javascript:alert(1)"};`, "123"); err == nil {
		t.Fatal("non-media link accepted")
	}
	if got := lujj31SameSite("https://lujj31.lol", "https://evil.example/vod/play/id/123"); got != "" {
		t.Fatal("cross-origin page accepted", got)
	}
	if got := lujj31SourceID("/index.php/vod/play/id/123/sid/1/nid/2.html"); got != "123" {
		t.Fatal("wrong route ID", got)
	}
	d := &Downloader{}
	if _, err := d.resolveLujj31Media(context.Background(), Task{DramaID: "lujj31:123", Chapter: Chapter{ID: "lujj31:999:1-1", PageURL: "https://lujj31.lol/index.php/vod/play/id/999/sid/1/nid/1.html"}}); err == nil {
		t.Fatal("other title's chapter accepted")
	}
}

func TestLujj31BlankHomepage(t *testing.T) {
	d := sourceFixtureDownloader(t, func(r *http.Request) (*http.Response, error) {
		return sourceFixtureResponse(r, http.StatusOK, ""), nil
	})
	_, err := d.fetchLujj31Categories(context.Background())
	if err == nil || !strings.Contains(err.Error(), "空白网页") {
		t.Fatalf("blank response must be actionable: %v", err)
	}
	doc, _ := html.Parse(strings.NewReader(`<div class="video-item"><a href="/index.php/vod/detail/id/55.html" title="测试影片">详情</a></div>`))
	if got := lujj31Cards(doc, "https://lujj31.lol"); len(got) != 1 || got[0].SourceID != "55" {
		t.Fatalf("fallback card: %+v", got)
	}
}

func TestLujj31OktvCSSPosterAndDetailCover(t *testing.T) {
	cards, err := html.Parse(strings.NewReader(`<html><body>
		<div class="myui-vodlist__box">
			<a href="/index.php/vod/detail/id/456.html" title="有封面影片"
			class="myui-vodlist__thumb" style="background-image: url('/upload/posters/456.jpg')"></a>
			<img src="/static/images/loading.gif">
		</div></body></html>`))
	if err != nil {
		t.Fatal(err)
	}
	list := lujj31Cards(cards, "https://lujj31.lol/index.php/vod/type/id/1.html")
	if len(list) != 1 || list[0].Cover != "https://lujj31.lol/upload/posters/456.jpg" {
		t.Fatalf("CSS poster not ported from oktv: %+v", list)
	}
	doc, err := html.Parse(strings.NewReader(`<html><head>
		<meta property="og:image" content="https://cdn.example.test/poster.jpg">
		</head><body><h1>有封面影片</h1>
		<div class="module-play-list-content">
			<a href="/index.php/vod/play/id/456/sid/1/nid/1.html">第1集</a>
		</div></body></html>`))
	if err != nil {
		t.Fatal(err)
	}
	detail, episodes, err := lujj31DetailFromHTML(doc, "https://lujj31.lol", "456")
	if err != nil || len(episodes) != 1 || detail.Cover != "https://cdn.example.test/poster.jpg" {
		t.Fatalf("detail OG cover not preserved: %+v episodes=%d err=%v", detail, len(episodes), err)
	}
}

func TestLujj31PlayerAcceptsOktvHTTPAndIframeB64(t *testing.T) {
	url, err := lujj31PlayerURL(`player_aaaa={"id":123,"encrypt":0,"url":"https://stream.example.test/token/play?id=123"};`, "123")
	if err != nil || url != "https://stream.example.test/token/play?id=123" {
		t.Fatalf("non-extension HTTP player rejected: %q %v", url, err)
	}
	decoded := "https://cdn.example.test/episode.m3u8"
	encoded := base64.StdEncoding.EncodeToString([]byte(decoded))
	url, err = lujj31PlayerURL(`<iframe src="https://player.example.test/player?url=`+encoded+`"></iframe>`, "123")
	if err != nil || url != decoded {
		t.Fatalf("iframe-base64 m3u8 lost: %q %v", url, err)
	}
	if got := lujj31Decrypt("%u7B2C%201%E9%9B%86", "1"); got != "第 1集" {
		t.Fatalf("macCMS JS escape() not decoded: %q", got)
	}
}

func TestLujj31OktvAlternateSIDPlayback(t *testing.T) {
	var requests []string
	d := sourceFixtureDownloader(t, func(r *http.Request) (*http.Response, error) {
		requests = append(requests, r.URL.Path)
		switch {
		case strings.Contains(r.URL.Path, "/sid/1/nid/1.html"):
			return sourceFixtureResponse(r, http.StatusOK, `player_data={"id":123,"encrypt":0,"url":"encrypted-token"};`), nil
		case strings.Contains(r.URL.Path, "/sid/2/nid/1.html"):
			return sourceFixtureResponse(r, http.StatusOK, `player_aaaa={"id":123,"encrypt":0,"url":"https://cdn.example.test/backup.m3u8"};`), nil
		default:
			return sourceFixtureResponse(r, http.StatusNotFound, ""), nil
		}
	})
	page := "https://lujj31.lol/index.php/vod/play/id/123/sid/1/nid/1.html"
	address, err := d.lujj31ResolveAlternative(context.Background(), page, "123")
	if err != nil || address != "https://cdn.example.test/backup.m3u8" {
		t.Fatalf("alternate oktv source line ignored: url=%q err=%v requests=%v", address, err, requests)
	}
	if len(requests) == 0 || !strings.Contains(requests[0], "/sid/2/") {
		t.Fatalf("first fallback did not use sid 2: %v", requests)
	}
}

func TestLujj31OktvCloudParserFallback(t *testing.T) {
	var sawFallback bool
	d := sourceFixtureDownloader(t, func(request *http.Request) (*http.Response, error) {
		switch {
		case request.URL.Host == "lujj31.lol" && request.URL.Path == "/static/player/m3u8.js":
			sawFallback = true
			return sourceFixtureResponse(request, http.StatusOK, `<iframe src="https://parser.example.test/parse?url="></iframe>`), nil
		case request.URL.Host == "parser.example.test" && request.URL.Query().Get("url") == "opaque-token":
			return sourceFixtureResponse(request, http.StatusOK, `var video_url="https://cdn.example.test/episode.m3u8";`), nil
		default:
			return sourceFixtureResponse(request, http.StatusNotFound, ""), nil
		}
	})
	page := "https://lujj31.lol/index.php/vod/play/id/123/sid/1/nid/1.html"
	address := d.lujj31CloudParse(context.Background(), `player_data={"url":"opaque-token","from":"site-line"};`, "opaque-token", page)
	if !sawFallback || address != "https://cdn.example.test/episode.m3u8" {
		t.Fatalf("oktv cloud parser fallback failed: %q fallback=%v", address, sawFallback)
	}
}

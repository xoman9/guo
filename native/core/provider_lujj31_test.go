package core

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestLujj31SourceRegistration(t *testing.T) {
	if !isHuangguoProviderSource(sourceLujj31) || canonicalProviderSource("lujj31.lol") != sourceLujj31 {
		t.Fatal("lujj31 must have a distinct registered source identity")
	}
	if providerSourceForURL("https://lujj31.lol/") != sourceLujj31 {
		t.Fatal("lujj31 host is not associated with its source")
	}
	address := "https://lujj31.lol/api.php/provide/vod/?ac=list"
	candidates := (&Downloader{cfg: defaultConfig()}).providerURLCandidates(address)
	if len(candidates) != 1 || candidates[0] != address {
		t.Fatalf("lujj31 request URL was unexpectedly rewritten: %v", candidates)
	}
}

func TestLujj31MacCMSFixture(t *testing.T) {
	d := sourceFixtureDownloader(t, func(request *http.Request) (*http.Response, error) {
		if request.URL.Host != "lujj31.lol" || request.URL.Path != "/api.php/provide/vod/" {
			t.Fatalf("unexpected lujj31 request: %s", request.URL)
		}
		query := request.URL.Query()
		var body string
		switch {
		case query.Get("ids") == "123":
			body = `{"list":[{"vod_id":123,"vod_name":"测试影片","vod_play_url":"第1集$https://media.test/1.m3u8#第2集$https://media.test/2.mp4"}]}`
		case query.Get("pg") == "2":
			body = `{"list":[{"vod_id":124,"vod_name":"第二部"}],"pagecount":3}`
		default:
			body = `{"class":[{"type_id":5,"type_name":"短剧"}],"list":[{"vod_id":123,"vod_name":"测试影片","vod_pic":"/poster.jpg"}],"pagecount":3}`
		}
		return sourceFixtureResponse(request, http.StatusOK, body), nil
	})
	categories, err := d.fetchLujj31Categories(context.Background())
	if err != nil || len(categories) != 1 || categories[0].ID != "5" {
		t.Fatalf("categories: %+v %v", categories, err)
	}
	items, more, err := d.fetchLujj31CatalogPage(context.Background(), 2, "5", "")
	if err != nil || !more || len(items) != 1 || items[0].ID != "lujj31:124" {
		t.Fatalf("catalog: %+v %v %v", items, more, err)
	}
	drama, chapters, err := d.fetchLujj31Detail(context.Background(), "123")
	if err != nil || drama.Source != sourceLujj31 || len(chapters) != 2 ||
		chapters[0].VideoURL != "https://media.test/1.m3u8" {
		t.Fatalf("detail: %+v %+v %v", drama, chapters, err)
	}
}

func TestLujj31RejectsNonMediaAndIncorrectIdentity(t *testing.T) {
	chapters := lujj31Chapters("123", "解析$javascript:alert(1)#下一集$https://cdn.test/2.mp4", "https://lujj31.lol/")
	if len(chapters) != 1 || !strings.HasPrefix(chapters[0].ID, "lujj31:123:") {
		t.Fatalf("unexpected chapters: %+v", chapters)
	}
	d := &Downloader{}
	_, err := d.resolveLujj31Media(context.Background(), Task{
		DramaID: "lujj31:123",
		Chapter: Chapter{ID: "lujj31:other:1", VideoURL: "https://cdn.test/1.mp4"},
	})
	if err == nil {
		t.Fatal("incorrect chapter identity was accepted")
	}
}

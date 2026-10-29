package core

import (
	"context"
	"net/http"
	"strings"
	"testing"

	nethtml "golang.org/x/net/html"
)

func TestLiuguoRegistrationAndLiveOnlyRouting(t *testing.T) {
	if !isHuangguoProviderSource(sourceLiuguo) ||
		canonicalProviderSource("www.llsp.me") != sourceLiuguo ||
		providerSourceForURL("https://www.llsp.me/vodtype/3.html") != sourceLiuguo {
		t.Fatal("Liuguo source or host is not registered")
	}
	address := "https://www.llsp.me/vodtype/3.html"
	candidates := (&Downloader{cfg: defaultConfig()}).providerURLCandidates(address)
	if len(candidates) != 1 || candidates[0] != address {
		t.Fatalf("Liuguo domain rehosted unexpectedly: %v", candidates)
	}
	if !validNativeCategory(sourceLiuguo, "23") || validNativeCategory(sourceLiuguo, "../etc") {
		t.Fatal("Liuguo category check failed")
	}
	if same := liuguoSameSite(liuguoBaseURL, "https://other.example/vodplay/1-1-1.html");same!="" {
		t.Fatal("cross-site play page accepted")
	}
}

func TestLiuguoOktvCategoryCatalogSearchAndPage(t *testing.T) {
	var paths []string
	d := sourceFixtureDownloader(t, func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		var body string
		switch r.URL.Path {
		case "/":
			body = `<html><body><a href="/vodtype/3.html">全部</a><a href="/vodtype/23.html">短劇</a><a href="/vodtype/24.html">電影</a></body></html>`
		case "/vodtype/3.html":
			body = `<html><body><div class="video-item">
				<a href="/vodtype/4.html">分類導航</a>
				<a href="/voddetail/901.html" title="&lt;b&gt;片名&lt;/b&gt; 甲">
				 <img data-original="/images/poster.jpg"></a>
				</div></body></html>`
		case "/vodtype/3-2.html":
			body = `<html><body><div class="media-card"><a href="/vodplay/902-1-1.html" title="影片乙">影片乙</a></div></body></html>`
		case "/vodsearch/合成-.html":
			body = `<html><body><div class="video-item"><a title="搜尋影片" href="/voddetail/903.html">搜尋影片</a></div></body></html>`
		default:
			return sourceFixtureResponse(r,http.StatusNotFound,"missing"),nil
		}
		return sourceFixtureResponse(r,http.StatusOK,body),nil
	})
	categories,err:=d.fetchLiuguoCategories(context.Background())
	if err!=nil||len(categories)!=2||categories[0].ID!="23"||categories[1].ID!="24" {
		t.Fatalf("category nav unexpected: %v %v",categories,err)
	}
	items,more,err:=d.fetchLiuguoCatalogPage(context.Background(),1,"","")
	if err!=nil||!more||len(items)!=1||items[0].ID!="liuguo:901"||items[0].Cover!="https://www.llsp.me/images/poster.jpg"||items[0].Title!="片名 甲"{
		t.Fatalf("catalog unexpected: %+v more=%v err=%v",items,more,err)
	}
	items,more,err=d.fetchLiuguoCatalogPage(context.Background(),2,"","")
	if err!=nil||!more||len(items)!=1||items[0].SourceID!="902"{
		t.Fatalf("page 2 unexpected: %+v more=%v err=%v",items,more,err)
	}
	items,more,err=d.fetchLiuguoCatalogPage(context.Background(),1,"","合成")
	if err!=nil||more||len(items)!=1||items[0].SourceID!="903"{
		t.Fatalf("search unexpected: %+v more=%v err=%v",items,more,err)
	}
	if len(paths)!=4 || paths[1]!="/vodtype/3.html" || paths[2]!="/vodtype/3-2.html" {
		t.Fatalf("not using oktv routes: %v",paths)
	}
}

func TestLiuguoOktvDetailCollapseAndCover(t *testing.T) {
	doc,err:=nethtml.Parse(strings.NewReader(`<html><head>
		<meta property="og:image" content="/poster.jpg"></head><body>
		<h1 class="page-title">合成影片甲</h1><div class="module-play-list-content">
			<a href="/vodplay/901-1-1.html">第1集</a>
			<a href="/vodplay/901-1-2.html">第2集</a>
			<a href="/vodplay/901-2-1.html">備用1</a>
			<a href="/vodplay/999-1-1.html">其他影片</a>
			<a href="https://evil.example/vodplay/901-1-3.html">錯誤網域</a>
		</div></body></html>`))
	if err!=nil {t.Fatal(err)}
	drama,chapters,err:=liuguoDetailFromHTML(doc,liuguoBaseURL,"901")
	if err!=nil||drama.Source!=sourceLiuguo||drama.Cover!="https://www.llsp.me/poster.jpg"||len(chapters)!=2{
		t.Fatalf("detail unexpected: %+v chapters=%+v err=%v",drama,chapters,err)
	}
	if chapters[1].PageURL!="https://www.llsp.me/vodplay/901-1-2.html"{
		t.Fatalf("wrong second episode: %+v",chapters[1])
	}
}

func TestLiuguoOktvPlaybackAndBackupSID(t *testing.T) {
	var requests []string
	d:=sourceFixtureDownloader(t,func(r *http.Request)(*http.Response,error){
		requests=append(requests,r.URL.Path)
		var body string
		switch r.URL.Path {
		case "/vodplay/901-1-1.html":
			body=`player_aaaa={"id":901,"encrypt":0,"url":"opaque-token","from":"none"};`
		case "/vodplay/901-2-1.html":
			body=`player_data={"id":901,"encrypt":0,"url":"https://media.test/episode.mp4"};`
		default:
			return sourceFixtureResponse(r,http.StatusNotFound,"not found"),nil
		}
		return sourceFixtureResponse(r,http.StatusOK,body),nil
	})
	page:="https://www.llsp.me/vodplay/901-2-1.html"
	address,err:=d.liuguoResolvePage(context.Background(),page,"901")
	if err!=nil||address!="https://media.test/episode.mp4"{
		t.Fatalf("Liuguo player unresolved: %q %v",address,err)
	}
	if _,err:=d.liuguoResolvePage(context.Background(),page,"902");err==nil{
		t.Fatal("wrong drama accepted")
	}
	alternates:=liuguoAlternateRoutes("https://www.llsp.me/vodplay/901-1-1.html")
	if len(alternates)==0||alternates[0]!="https://www.llsp.me/vodplay/901-2-1.html"{
		t.Fatalf("alternate SID routes invalid: %v",alternates)
	}
	if len(requests)!=1 {t.Fatalf("unexpected request count: %v",requests)}
}

func TestLiuguoNonCacheCatalog(t *testing.T) {
	var calls int
	engine:=sourceFixtureEngine(t,func(r *http.Request)(*http.Response,error){
		if r.URL.Path!="/vodtype/3.html"{t.Fatalf("unexpected live path: %s",r.URL.Path)}
		calls++
		return sourceFixtureResponse(r,http.StatusOK,`<html><body><div class="video-item"><a href="/voddetail/901.html" title="影片甲">影片甲</a></div></body></html>`),nil
	})
	for i:=0;i<2;i++{
		result,err:=engine.nativeCatalog(context.Background(),nativeInput{Source:sourceLiuguo,Page:1})
		if err!=nil||len(result.Items)!=1||result.Items[0].SourceID!="901"{
			t.Fatalf("live result %d: %+v err=%v",i,result,err)
		}
	}
	if calls!=2 {t.Fatalf("catalog not fetched live: requests=%d",calls)}
	if got:=engine.nativeCached(sourceLiuguo);len(got.Items)!=0{
		t.Fatalf("live catalog unexpectedly persisted: %+v",got)
	}
}

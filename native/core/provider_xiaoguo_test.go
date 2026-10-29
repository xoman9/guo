package core

import (
 "context"
 "net/http"
 "strings"
 "testing"

 nethtml "golang.org/x/net/html"
)

// Representative, non-explicit HTML structure from the user's 2026-10-29
// yese.co home-page source. The site renders both URL spellings for categories
// and uses /vod/play/id/<id>/sid/<line>/nid/<episode>/ with lazy-load images.
const xiaoguoHomeFixture = `<!doctype html><html><body>
<nav class="app-nav"><div class="container"><div class="row">
 <a class="h5 text-light" href="/vod/show/class/中文字幕/id/1.html"> 中文字幕 </a>
 <a class="h5 text-light" href="/vod/show/class/卡通动画/id/1.html"> 卡通动画 </a>
 <a class="h5 text-light" href="/vod/show/class/亚洲情色/id/1.html"> 亚洲情色 </a>
 <a href="/vod/show/class/中文字幕/id/1.html"> 中文字幕 </a>
</div></div></nav>
<header><a href="/vod/show/id/1.html">最新上傳</a></header>
<section><div class="more">
 <a href="/index.php/vod/show/class/中文字幕/id/1.html">更多</a>
</div><div class="col-6 col-sm-4 col-lg-3"><div class="video-img-box"><div class="img-box">
<a href="/vod/play/id/306091/sid/1/nid/1/"><img class="lazyload"
 src="/template/yese/images/placeholder.jpg"
 data-src="https://example.test/20261026/cover.jpg" title="範例影片" alt="範例影片"></a>
</div></div></div></section>
</body></html>`

func TestXiaoguoCategoriesFromActualHomeStructure(t *testing.T){
 doc,err:=nethtml.Parse(strings.NewReader(xiaoguoHomeFixture))
 if err!=nil{t.Fatal(err)}
 cats:=xiaoguoCategoriesFromHTML(doc)
 if len(cats)!=3 {t.Fatalf("expected only 3 classes, got %+v",cats)}
 for _,c:=range cats{
  if c.Name=="全部" || c.Name=="更多" || c.Name=="最新上傳"{
   t.Fatalf("navigation leaked into classes: %+v",cats)
  }
  if !validNativeCategory(sourceXiaoguo,c.ID){t.Fatalf("category failed native validation: %+v",c)}
  path,err:=xiaoguoCategoryURL(c.ID,1)
  if err!=nil || !strings.HasPrefix(path,"/vod/show/class/") || !strings.HasSuffix(path,"/id/1.html"){
   t.Fatalf("class route is wrong %q: %v",path,err)
  }
 }
 a:=xiaoguoCategoryID("/vod/show/class/中文字幕/id/1.html")
 b:=xiaoguoCategoryID("/index.php/vod/show/class/中文字幕/id/1.html")
 if a==""||a!=b{t.Fatalf("duplicate category has mismatched keys %q %q",a,b)}
 if xiaoguoCategoryID("/vod/show/id/1.html")!=""{t.Fatal("all/newest leaked as a class")}
 if validNativeCategory(sourceXiaoguo,"p:broken") || validNativeCategory(sourceXiaoguo,"../invalid"){
  t.Fatal("invalid category accepted")
 }
}

func TestXiaoguoHTMLMovieCardsAndLazyCover(t *testing.T){
 doc,err:=nethtml.Parse(strings.NewReader(xiaoguoHomeFixture))
 if err!=nil{t.Fatal(err)}
 cards:=lujj31Cards(doc,xiaoguoBaseURL)
 if len(cards)!=1||cards[0].SourceID!="306091" || cards[0].Cover!="https://example.test/20261026/cover.jpg"{
  t.Fatalf("failed to recognize /vod/play/id/.../sid/.../nid and lazy cover: %+v",cards)
 }
}

func TestXiaoguoLiveSourceRoutes(t *testing.T){
 var requestPaths []string
 d:=sourceFixtureDownloader(t,func(r *http.Request)(*http.Response,error){
  requestPaths=append(requestPaths,r.URL.Path)
  switch r.URL.Path{
  case "/":return sourceFixtureResponse(r,http.StatusOK,xiaoguoHomeFixture),nil
  case "/vod/show/class/中文字幕/id/1.html":
   return sourceFixtureResponse(r,http.StatusOK,xiaoguoHomeFixture),nil
  case "/vod/show/class/中文字幕/id/1/page/2.html":
   return sourceFixtureResponse(r,http.StatusOK,xiaoguoHomeFixture),nil
  default:return sourceFixtureResponse(r,http.StatusNotFound,""),nil
  }
 })
 categories,err:=d.fetchXiaoguoCategories(context.Background())
 if err!=nil||len(categories)!=3{t.Fatalf("site categories %v %v",categories,err)}
 items,more,err:=d.fetchXiaoguoCatalogPage(context.Background(),1,categories[0].ID,"")
 if err!=nil||!more||len(items)!=1||items[0].ID!="xiaoguo:306091"{
  t.Fatalf("catalog wrong items=%+v more=%v err=%v",items,more,err)
 }
 _,_,err=d.fetchXiaoguoCatalogPage(context.Background(),2,categories[0].ID,"")
 if err!=nil{t.Fatalf("category second page path missing: %v (%v)",err,requestPaths)}
 if len(requestPaths)!=3 || requestPaths[1]!="/vod/show/class/中文字幕/id/1.html" || requestPaths[2]!="/vod/show/class/中文字幕/id/1/page/2.html"{
  t.Fatalf("unexpected route requests %v",requestPaths)
 }
}

func TestXiaoguoAdultCatalogBypassesDiskCache(t *testing.T){
 var requests int
 engine:=sourceFixtureEngine(t,func(r *http.Request)(*http.Response,error){
  requests++
  if r.URL.Path!="/index.php/vod/show/id/1.html"{t.Fatalf("unexpected Xiaoguo route %s",r.URL.Path)}
  return sourceFixtureResponse(r,http.StatusOK,xiaoguoHomeFixture),nil
 })
 for i:=0;i<2;i++{
  result,err:=engine.nativeCatalog(context.Background(),nativeInput{Source:sourceXiaoguo,Page:1})
  if err!=nil||len(result.Items)!=1||result.Items[0].SourceID!="306091"{
   t.Fatalf("live catalog %d: %+v %v",i,result,err)
  }
 }
 if requests!=2{t.Fatalf("Xiaoguo catalog was cached, requests=%d",requests)}
 if result:=engine.nativeCached(sourceXiaoguo);len(result.Items)!=0{
  t.Fatalf("Xiaoguo unexpectedly saved a disk catalog %+v",result)
 }
}

package core

import (
 "context"
 "net/http"
 "strings"
 "testing"
 "sync"
 "strconv"

 nethtml "golang.org/x/net/html"
)

func TestApk391NativeIdentityAndIsolation(t *testing.T) {
 if len(apk391Sources)!=43 || len(apk391ByID)!=43 {
  t.Fatalf("expected 43 registered sources after removing two, got %d / %d",len(apk391Sources),len(apk391ByID))
 }
 for _,spec:=range apk391Sources {
  if !isHuangguoProviderSource(spec.ID) || canonicalProviderSource(spec.ID)!=spec.ID{
   t.Fatalf("registered source not recognized: %+v",spec)
  }
  if apk391BaseURL(spec.ID)=="" || apk391SourceForHost(spec.Host)!=spec.ID{
   t.Fatalf("APK source registry host mapping lost: %+v",spec)
  }
  if providerSourceForURL("https://"+spec.Host+"/")!=spec.ID {
   t.Fatalf("global host/source mapping lost: %+v",spec)
  }
 }
 for _,id:=range []string{"lujj31","liuguo","xiaoguo","hongguo","chaoguo","chengguo"}{
  if _,ok:=apk391Spec(id);ok{t.Fatalf("completed provider overwritten: %s",id)}
 }
 if canonicalProviderSource("hsck.net")!="cangguo"||canonicalProviderSource("javbus.com")!="baguo" {
  t.Fatal("source aliases cannot replace native IDs")
 }
 total := 0
 groups := 0
 for _,rows:=range apk391StaticCategoryFallback {
   total+=len(rows)
   if len(rows)>0{groups++}
 }
 if total<420 || groups<18 { t.Fatalf("missing active APK static classifications: %d rows, %d groups",total,groups)}
 if len(apk391StaticCategoryFallback["luguo"])<5 {
  t.Fatal("historical categories missing")
 }
}

func TestApk391MacCMSCardsSameOriginAndDetail(t *testing.T) {
 source:="yanguo"
 fixture:=`<html><head><meta property="og:image" content="/poster.png"></head><body>
 <div class="module-item"><a href="/voddetail/901.html" title="影片甲"><img data-original="/cover.jpg"></a></div>
 <div class="module-item"><a href="https://foreign.invalid/voddetail/909.html" title="其它站点">其它站点</a></div>
 <div class="module-play-list-content">
 <a href="/vodplay/901-1-1.html">第1集</a><a href="/vodplay/901-1-2.html">第2集</a>
 <a href="/vodplay/902-1-1.html">错误影片</a></div></body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(fixture))
 if err!=nil{t.Fatal(err)}
 items:=apk391Cards(doc,source)
 if len(items)==0 || items[0].ID!="yanguo:901" {
  t.Fatalf("card not recognized: %+v",items)
 }
 if items[0].Cover!="https://xqxq1.cc/cover.jpg"{
  t.Fatalf("cover is not source-local: %+v",items[0])
 }
 d:=sourceFixtureDownloader(t,func(r *http.Request)(*http.Response,error){
  if r.URL.Path!="/index.php/vod/detail/id/901.html"{
   return sourceFixtureResponse(r,http.StatusNotFound,"missing"),nil
  }
  return sourceFixtureResponse(r,http.StatusOK,fixture),nil
 })
 drama,chapters,err:=d.fetchApk391Detail(context.Background(),source,"901")
 if err!=nil || drama.Source!="yanguo" || len(chapters)!=2 ||
  chapters[1].PageURL!="https://xqxq1.cc/vodplay/901-1-2.html" {
  t.Fatalf("detail unexpected: %+v %+v %v",drama,chapters,err)
 }
 if apk391SameSite(source,"https://foreign.invalid/voddetail/901.html")!=""{
  t.Fatal("cross-origin detail accepted")
 }
}

func TestApk391ChengguoRealSlugAndCategoryCards(t *testing.T) {
 if !apk391ValidCategory("chengguodj","yuanchuang")||!apk391ValidCategory("chengguodj","manju"){t.Fatal("missing verified Chengguo categories")}
 if apk391ValidCategory("chengguodj","fake"){t.Fatal("unknown category accepted")}
 htmlText:=`<html><body><section><article data-xpch="card-drama"><a href="/drama/dj-874b37a0"><img src="https://images.example.org/cover.jpg"></a><h3>正常短劇片名</h3></article></section></body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(htmlText));if err!=nil{t.Fatal(err)}
 cards:=chengguoCards(doc)
 if len(cards)!=1||cards[0].SourceID!="dj-874b37a0"||cards[0].Title!="正常短劇片名"{
  t.Fatalf("unexpected Chengguo catalog cards: %+v",cards)
 }
}

func TestApk391CatalogDoesNotDisplayCategoryNamesAsMovies(t *testing.T) {
 source:="yanguo"
 fake:=`<html><body>
 <nav><a href="/voddetail/111.html">国产自拍</a></nav>
 <div class="module-item"><a href="/voddetail/112.html" title="国产传媒">国产传媒</a></div>
 <div class="module-item"><a href="/voddetail/113.html" title="探花系列">探花系列</a></div>
 <div class="module-item"><a href="/voddetail/114.html" title="人妻熟女">人妻熟女</a></div>
 <div class="module-play-list-content"><a href="/vodplay/999-1-1.html" title="第1集"><img src="/irrelevant.jpg"></a></div>
 <div class="module-item"><a href="/voddetail/901.html" title="国产传媒">
   <img data-original="/posters/901.jpg" alt="真正影片的片名"></a></div>
 </body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(fake));if err!=nil{t.Fatal(err)}
 cards:=apk391Cards(doc,source)
 if len(cards)!=1 || cards[0].ID!="yanguo:901" ||
  cards[0].Title!="真正影片的片名" ||
  cards[0].Cover!="https://xqxq1.cc/posters/901.jpg" {
  t.Fatalf("taxonomy entries leaked into movies: %+v",cards)
 }
 for _,c:=range cards {
  for _,category:=range []string{"国产自拍","国产传媒","探花系列","人妻熟女"} {
   if c.Title==category {t.Fatalf("category name used as video title: %+v",c)}
  }
 }
}

func TestApk391CatalogPageKeepsActualPostersAfterFiltering(t *testing.T) {
 d:=sourceFixtureDownloader(t,func(r *http.Request)(*http.Response,error){
  if r.URL.Path!="/" {return sourceFixtureResponse(r,http.StatusNotFound,"missing"),nil}
  body:=`<html><body><nav><a href="/voddetail/111.html">国产自拍</a></nav>
  <section class="module-item"><a href="/voddetail/112.html">国产传媒</a></section>
  <section class="module-item"><a href="/voddetail/901.html"><img alt="真正影片" data-src="/cover901.jpg"></a></section><a rel="next" href="/index.php/vod/show/id/1/page/2.html">下一页</a>
  </body></html>`
  return sourceFixtureResponse(r,http.StatusOK,body),nil
 })
 items,more,err:=d.fetchApk391CatalogPage(context.Background(),"yanguo",1,"","")
 if err!=nil||len(items)!=1||items[0].Title!="真正影片"||!more {
  t.Fatalf("expected real movie only: %+v more=%v err=%v",items,more,err)
 }
}


func TestApk391OnDemandDetailEnrichmentCorrectsCategoryLikeTitles(t *testing.T) {
 var mu sync.Mutex
 calls:=map[string]int{}
 d:=sourceFixtureDownloader(t,func(req *http.Request)(*http.Response,error){
  mu.Lock()
  calls[req.URL.Path]++
  mu.Unlock()
  switch req.URL.Path {
  case "/":
   return sourceFixtureResponse(req,http.StatusOK,`<html><body>
   <div class="module-item"><a href="/voddetail/901.html">国产自拍</a></div>
   <div class="module-item"><a href="/voddetail/902.html">国产传媒</a></div>
   <div class="module-item"><a href="/voddetail/903.html">探花系列</a></div>
   <div class="module-item"><a href="/voddetail/904.html">人妻熟女</a></div>
   <a rel="next" href="/index.php/vod/show/id/1/page/2.html">下一页</a>
   </body></html>`),nil
  case "/voddetail/901.html":
   return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/posters/901.jpg"></head><body><h1>影片甲的真正片名</h1></body></html>`),nil
  case "/voddetail/902.html":
   return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/posters/902.jpg"></head><body><h1>影片乙的真正片名</h1></body></html>`),nil
  case "/voddetail/903.html":
   return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/posters/903.jpg"></head><body><h1>影片丙的真正片名</h1></body></html>`),nil
  case "/voddetail/904.html":
   return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/posters/904.jpg"></head><body><h1>影片丁的真正片名</h1></body></html>`),nil
  default:
   return sourceFixtureResponse(req,http.StatusNotFound,"missing"),nil
  }
 })
 items,more,err:=d.fetchApk391CatalogPage(context.Background(),"yanguo",1,"","")
 if err!=nil||!more||len(items)!=4 {
  t.Fatalf("expected four real detail-enriched videos, got %+v more=%v err=%v",items,more,err)
 }
 names:=[]string{"影片甲的真正片名","影片乙的真正片名","影片丙的真正片名","影片丁的真正片名"}
 for i,v:=range items {
  if v.Title!=names[i]||v.Cover==nil||v.Cover=="" {
   t.Fatalf("item %d still displays taxonomy or placeholder: %+v",i,v)
  }
 }
 mu.Lock()
 defer mu.Unlock()
 if calls["/"]!=1||len(calls)!=5 {
  t.Fatalf("unexpected detail prefetch beyond four uncertain items: %+v",calls)
 }
}


func TestApk391LuguoCategoryPosterResolvesActualTitle(t *testing.T) {
 d:=sourceFixtureDownloader(t,func(req *http.Request)(*http.Response,error){
  switch req.URL.Path {
  case "/":
   return sourceFixtureResponse(req,http.StatusOK,`<html><body>
    <div class="module-item"><a title="国产传媒" href="/voddetail/601.html">
     <img data-src="/covers/601.jpg" title="国产传媒"></a></div>
    <div class="module-item"><a title="探花系列" href="/voddetail/602.html">
     <img data-src="/covers/602.jpg"></a></div>
    </body></html>`),nil
  case "/voddetail/601.html":
   return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/covers/601.jpg"></head>
    <body><h1>露果電影甲 真正片名</h1></body></html>`),nil
  case "/voddetail/602.html":
   return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/covers/602.jpg"></head>
    <body><h1>露果電影乙 真正片名</h1></body></html>`),nil
  default: return sourceFixtureResponse(req,http.StatusNotFound,"missing"),nil
  }
 })
 items,more,err:=d.fetchApk391CatalogPage(context.Background(),"luguo",1,"","")
 if err!=nil||more||len(items)!=2 {
  t.Fatalf("Luguo catalog title enrichment returned %+v more=%v err=%v",items,more,err)
 }
 if items[0].Title!="露果電影甲 真正片名"||items[1].Title!="露果電影乙 真正片名" {
  t.Fatalf("Luguo showed category names as video titles: %+v",items)
 }
}

func TestApk391IgnoreCategoryAndPaginationIDs(t *testing.T) {
 for _,path:=range []string{
  "/index.php/vod/show/id/1/page/2.html", "/vod/type/id/22.html",
  "/vodtype/3-2.html", "/vodshow/id/12.html"}{
  if got:=apk391VideoID(path);got!="" {
   t.Fatalf("category or paging URL treated as video: %s => %s",path,got)
  }
 }
 for _,path:=range []string{"/voddetail/601.html","/vodplay/601-1-1.html","/vod/play/id/601/sid/1/nid/1/"}{
  if got:=apk391VideoID(path);got!="601"{
   t.Fatalf("real video no longer recognized: %s => %s",path,got)
  }
 }
 doc,err:=nethtml.Parse(strings.NewReader(`<html><body>
  <a href="/index.php/vod/show/id/1/page/2.html">下一页</a>
 </body></html>`))
 if err!=nil{t.Fatal(err)}
 if !apk391HasNextPage(doc,1)||apk391HasNextPage(doc,2){
  t.Fatal("pagination successor detection returned wrong result")
 }
}

func TestYanguoNativeMergeRetainsListPoster(t *testing.T) {
 before:=nativeDrama{ID:"yanguo:601",Source:"yanguo",Title:"正確片名",Cover:"https://xqxq1.cc/movie601.jpg"}
 after:=nativeDrama{ID:"yanguo:601",Source:"yanguo",Title:"更新片名",Cover:"https://xqxq1.cc/fallback.png"}
 merged:=mergeNativeDrama(before,after)
 if merged.Cover!=before.Cover||merged.Title!="更新片名"{
  t.Fatalf("playback detail replaced Yanguo list poster: %+v",merged)
 }
 empty:=mergeNativeDrama(nativeDrama{ID:"yanguo:602",Source:"yanguo"},nativeDrama{ID:"yanguo:602",Source:"yanguo",Cover:"https://xqxq1.cc/602.jpg"})
 if empty.Cover!="https://xqxq1.cc/602.jpg"{
  t.Fatalf("missing original cover blocked new detail poster: %+v",empty)
 }
}

func TestApk391LiguoAndLuguoUsePerVideoDetails(t *testing.T) {
 for _,source:=range []string{"luguo","liguo"} {
  t.Run(source,func(t *testing.T){
   base:=apk391BaseURL(source)
   if base==""{t.Fatal("missing source")}
   d:=sourceFixtureDownloader(t,func(req *http.Request)(*http.Response,error){
    switch req.URL.Path{
    case "/":
     // One shared parent, duplicated misleading poster, category and
     // duration labels: no metadata may be copied from a neighboring card.
     return sourceFixtureResponse(req,http.StatusOK,`<html><body>
      <div class="film-row">
       <a href="/voddetail/601.html" title="热门"><img src="/wrong-shared.jpg"></a>
       <a href="/voddetail/602.html" title="01:57:57"><img src="/wrong-shared.jpg"></a>
       <a href="/voddetail/603.html"><img src="/wrong-shared.jpg"></a>
      </div></body></html>`),nil
    case "/voddetail/601.html":
     return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/right601.jpg"></head><body><h1>影片一的真實片名</h1></body></html>`),nil
    case "/voddetail/602.html":
     return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/right602.jpg"></head><body><h1>影片二的真實片名</h1></body></html>`),nil
    case "/voddetail/603.html":
     return sourceFixtureResponse(req,http.StatusOK,`<html><head><meta property="og:image" content="/right603.jpg"></head><body><h1>影片三的真實片名</h1></body></html>`),nil
    }
    return sourceFixtureResponse(req,http.StatusNotFound,"missing"),nil
   })
   items,more,err:=d.fetchApk391CatalogPage(context.Background(),source,1,"","")
   if err!=nil||more||len(items)!=3{
    t.Fatalf("wanted 3 detail-backed items, got %+v more=%v err=%v",items,more,err)
   }
   wantTitles:=[]string{"影片一的真實片名","影片二的真實片名","影片三的真實片名"}
   for i,item:=range items {
    wantCover:=base+"/right"+strconv.Itoa(601+i)+".jpg"
    if item.SourceID!=strconv.Itoa(601+i)||item.Title!=wantTitles[i]||item.Cover!=wantCover{
     t.Fatalf("%s film %d wrongly bound metadata: %+v want title %s cover %s",source,i,item,wantTitles[i],wantCover)
    }
   }
  })
 }
}
func TestApk391RejectsDurationAndPopularTagsAsFilmTitles(t *testing.T){
 for _,title:=range []string{"热门","熱門","01:57:57","02:09:21","29:33","15:07","国产传媒"}{
  if !apk391BadTitle("liguo",title){t.Fatalf("unfiltered bogus film title %q",title)}
 }
}

func TestApk391Original391FilmDetailRoutes(t *testing.T) {
 for _,source:=range []string{"luguo","liguo"} {
  got:=apk391DetailRoutes(source,"602")
  if len(got)!=2||got[0]!="/vodplay/602-1-1/"||got[1]!="/vodplay/602-1-1.html" {
   t.Fatalf("%s must use original 0.2.391 detail routes: %+v",source,got)
  }
 }
 other:=apk391DetailRoutes("yanguo","602")
 if len(other)!=4||other[0]!="/index.php/vod/detail/id/602.html" {
  t.Fatalf("unrelated channel routes changed: %+v",other)
 }
}

func TestApk391LuguoAndLiguoReferenceHosts(t *testing.T) {
 cases:=map[string]string{"luguo":"https://luyitian.com","liguo":"https://ririlu.cc"}
 for source,want:=range cases {
  if got:=apk391BaseURL(source);got!=want {
   t.Fatalf("%s source host = %q, want %q",source,got,want)
  }
  if got:=apk391SameSite(source,want+"/vodplay/601-1-1/");got!=want+"/vodplay/601-1-1/" {
   t.Fatalf("%s canonical same-site URL invalid: %q",source,got)
  }
 }
}

func TestApk391LiguoWntheme27ActualCardStructure(t *testing.T) {
 fixture:=`<html><body><div class="list"><div class="list-title"><a href="/vodshow/28-----------/" class="more">更多</a></div><div class="list-content">
 <a href="/vodplay/288007-1-1/" class="card el-col"><div class="card-box"><div class="el-image"><img class="lazyload" src="/template/wntheme27/asset/img/lazyload.gif" data-src="https://img.example.org/actual.jpg"><div class="play-time">01:58:55</div></div><div class="desc">MOND-259 真正影片標題</div></div></a>
 <a href="/vodplay/288008-1-1/" class="card el-col"><div class="card-box"><div class="desc">只有文字無封面</div></div></a>
 </div></div></body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(fixture));if err!=nil{t.Fatal(err)}
 cards:=apk391LiguoCards(doc)
 if len(cards)!=1||cards[0].SourceID!="288007"||cards[0].Title!="MOND-259 真正影片標題"||cards[0].Cover!="https://img.example.org/actual.jpg"{
  t.Fatalf("wrong original wntheme27 movie metadata: %+v",cards)
 }
 if apk391Cards(doc,"liguo")[0].Title!=cards[0].Title{t.Fatal("liguo path did not use dedicated cards")}
}

func TestApk391LuguoAndLiguoDetailPreserveCatalogTitle(t *testing.T) {
 for _,source:=range []string{"luguo","liguo"} {
  t.Run(source,func(t *testing.T){
   const id="987654321"
   const expected="MIDV-586 真正的完整片名"
   apk391RememberTitle(source,id,expected)
   t.Cleanup(func(){apk391CatalogNames.Delete(source+":"+id)})
   d:=sourceFixtureDownloader(t,func(req *http.Request)(*http.Response,error){
    if req.URL.Path=="/vodplay/"+id+"-1-1/" {
     return sourceFixtureResponse(req,http.StatusOK,`<html><body><h1>影片 987654321</h1><script>var player_aaaa={"flag":"hls","encrypt":0,"url":"https://example.org/video.m3u8"};</script><a href="/vodplay/987654321-1-1/">第1集</a></body></html>`),nil
    }
    return sourceFixtureResponse(req,http.StatusNotFound,"missing"),nil
   })
   drama,_,err:=d.fetchApk391Detail(context.Background(),source,id)
   if err!=nil{t.Fatalf("fetch detail failed: %v",err)}
   if drama.Title!=expected||drama.Name!=expected{t.Fatalf("%s catalog title overwritten: %+v",source,drama)}
  })
 }
}

func TestApk391ZaoguoDirectoryIsCategories(t *testing.T) {
 cats:=apk391StaticCategoryFallback["zaoguo"]
 if len(cats)!=16{t.Fatalf("expected 16 directory categories, got %d",len(cats))}
 if apk391BaseURL("zaoguo")!="https://www.yhsp9.homes"{t.Fatal("Zaoguo host mismatch")}
 if !apk391ValidCategory("zaoguo","24") {t.Fatal("missing directory category 24")}
 const id="1197978"
 fixture:=`<html><body><div class="panel-heading"><h3 class="panel-title">目录</h3></div><ul class="navigation"><li><a href="/yhsp/index.php/vod/type/id/22.html">美女直播</a></li></ul>
 <ul class="videos"><li><div class="video"><a class="thumbnail" href="/yhsp/index.php/vod/play/id/1197978/sid/1/nid/1.html" title="有效片名"><div class="video-thumb"><img src="https://img.example.org/film.jpg" alt="有效片名"></div><span class="video-title">有效片名</span><span class="video-overlay">精品推荐</span></a></div></li></ul>
 <ul class="pagination"><li><a href="/yhsp/index.php/vod/type/id/22/page/2.html">下一页</a></li></ul></body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(fixture));if err!=nil{t.Fatal(err)}
 rows:=apk391ZaoguoCards(doc)
 if len(rows)!=1||rows[0].SourceID!=id||rows[0].Title!="有效片名"||rows[0].Cover!="https://img.example.org/film.jpg" {
  t.Fatalf("wrong zaoguo video card: %+v",rows)
 }
 if !apk391HasNextPage(doc,1){t.Fatal("actual /page/2.html not identified")}
 paths:=apk391DetailRoutes("zaoguo",id)
 if len(paths)!=1||paths[0]!="/yhsp/index.php/vod/play/id/1197978/sid/1/nid/1.html"{t.Fatalf("unexpected zaoguo detail route %+v",paths)}
}

func TestApk391JuguoVideo7CardsAndCategories(t *testing.T) {
 if apk391BaseURL("juguo")!="https://bulunhufait.buzz" {t.Fatal("stale Juguo domain")}
 if !apk391ValidCategory("juguo","372")||!apk391ValidCategory("juguo","423"){t.Fatal("missing observed video7 categories")}
 fixture:=`<html><body><div class="menu"><a href="/vodtype/372/">欧美性爱</a></div>
 <div class="mod index-list"><div class="row col5 clearfix">
 <dl><dt class="preview-item"><a href="/voddetail/1979090/"><img class="nature lazy" data-original="https://img.example.org/real.jpg" src="/template/video7/static/images/loading.svg"></a></dt><dd><a href="/voddetail/1979090/"><h3>原有完整片名</h3></a></dd></dl>
 <dl><dt class="preview-item"><a href="javascript:void(0)"><img data-original="/fake.jpg"></a></dt><dd><a href="javascript:void(0)"><h3>廣告</h3></a></dd></dl>
 </div></div></body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(fixture));if err!=nil{t.Fatal(err)}
 cards:=apk391JuguoCards(doc)
 if len(cards)!=1||cards[0].SourceID!="1979090"||cards[0].Title!="原有完整片名"||cards[0].Cover!="https://img.example.org/real.jpg" {
  t.Fatalf("incorrect video7 cards: %+v",cards)
 }
 if got:=apk391DetailRoutes("juguo","1979090")[0];got!="/voddetail/1979090/"{t.Fatalf("wrong Juguo detail URL %q",got)}
}

func TestApk391ZaoguoCatalogMetadataSurvivesDetail(t *testing.T) {
 const source="zaoguo"
 const id="987654321"
 const title="CEMD749 原始影片片名"
 const cover="https://img.example.org/zaoguo-cover.jpg"
 apk391RememberTitle(source,id,title)
 apk391RememberCover(source,id,cover)
 t.Cleanup(func(){apk391CatalogNames.Delete(source+":"+id);apk391CatalogCovers.Delete(source+":"+id)})
 d:=sourceFixtureDownloader(t,func(req *http.Request)(*http.Response,error){
  if req.URL.Path=="/yhsp/index.php/vod/play/id/"+id+"/sid/1/nid/1.html"{
   return sourceFixtureResponse(req,http.StatusOK,`<html><body><h1>影片 987654321</h1><script>var player_aaaa={"flag":"hls","encrypt":0,"url":"https://example.org/video.m3u8"};</script><a href="/yhsp/index.php/vod/play/id/987654321/sid/1/nid/1.html">第1集</a></body></html>`),nil
  }
  return sourceFixtureResponse(req,http.StatusNotFound,"missing"),nil
 })
 drama,_,err:=d.fetchApk391Detail(context.Background(),source,id)
 if err!=nil{t.Fatalf("fetch detail: %v",err)}
 if drama.Title!=title||drama.Cover!=cover{t.Fatalf("catalog info lost: %+v",drama)}
}

func TestApk391NingguoPosterMatchesFilmAnchor(t *testing.T) {
 fixture:=`<html><body><div class="row"><div class="grid">
 <a href="/voddetail/987601.html" title="第一部"><img data-src="https://img.example.org/first.jpg"></a>
 <a href="/voddetail/987602.html" title="第二部"><img data-src="https://img.example.org/second.jpg"></a>
 </div></div></body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(fixture));if err!=nil{t.Fatal(err)}
 if apk391RequiresDetail("ningguo"){t.Fatal("do not change valid Ningguo title/play flow")}
 var anchors []*nethtml.Node
 for _,a:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="a"}){
  if apk391VideoID(apk391SameSite("ningguo",providerHTMLAttr(a,"href")))!=""{anchors=append(anchors,a)}
 }
 if len(anchors)!=2{t.Fatalf("wrong film count: %d",len(anchors))}
 for i,want:=range []string{"https://img.example.org/first.jpg","https://img.example.org/second.jpg"}{
  got:=apk391ScopedPoster(anchors[i],"ningguo",apk391SameSite("ningguo",providerHTMLAttr(anchors[i],"href")))
  if got!=want{t.Fatalf("film %d poster %q != %q",i,got,want)}
 }
}

func TestApk391NingguoPosterRemainsAfterDetailNavigation(t *testing.T) {
 const source="ningguo"
 const id="987654321"
 const cover="https://img.example.org/correct-poster.jpg"
 apk391RememberCover(source,id,cover)
 t.Cleanup(func(){apk391CatalogCovers.Delete(source+":"+id)})
 d:=sourceFixtureDownloader(t,func(req *http.Request)(*http.Response,error){
  if req.URL.Path=="/index.php/vod/detail/id/"+id+".html" {
   return sourceFixtureResponse(req,http.StatusOK,`<html><body><h1>已確認的影片</h1><img src="https://img.example.org/other-film.jpg"><script>var player_aaaa={"flag":"hls","encrypt":0,"url":"https://example.org/video.m3u8"};</script><a href="/index.php/vod/play/id/987654321/sid/1/nid/1.html">播放</a></body></html>`),nil
  }
  return sourceFixtureResponse(req,http.StatusNotFound,"missing"),nil
 })
 drama,_,err:=d.fetchApk391Detail(context.Background(),source,id)
 if err!=nil{t.Fatalf("detail unavailable: %v",err)}
 if drama.Cover!=cover||drama.CoverURL!=cover{t.Fatalf("detail replaced catalog poster: %+v",drama)}
}

func TestApk391MangguoDetailRetainsCatalogCover(t *testing.T) {
 const source="mangguo"
 const id="987654321"
 const poster="https://img.example.org/mangguo-correct.jpg"
 apk391RememberCover(source,id,poster)
 t.Cleanup(func(){apk391CatalogCovers.Delete(source+":"+id)})
 d:=sourceFixtureDownloader(t,func(req *http.Request)(*http.Response,error){
  if req.URL.Path=="/index.php/vod/detail/id/"+id+".html"{
   return sourceFixtureResponse(req,http.StatusOK,`<html><body><h1>正确影片</h1><img src="https://img.example.org/wrong-recommendation.jpg"><a href="/index.php/vod/play/id/987654321/sid/1/nid/1.html">第1集</a></body></html>`),nil
  }
  return sourceFixtureResponse(req,http.StatusNotFound,"not found"),nil
 })
 movie,_,err:=d.fetchApk391Detail(context.Background(),source,id)
 if err!=nil{t.Fatal(err)}
 if movie.Cover!=poster||movie.CoverURL!=poster{t.Fatalf("incorrect cover after entering detail: %+v",movie)}
}

func TestApk391XiongguoUserConfirmedHost(t *testing.T) {
 spec,ok:=apk391Spec("xiongguo")
 if !ok||spec.Kind!="api"||spec.Host!="www.xm66xm.com"{
  t.Fatalf("Xiongguo must retain dedicated API type and confirmed host: %+v",spec)
 }
 if got:=apk391BaseURL("xiongguo");got!="https://www.xm66xm.com"{
  t.Fatalf("unexpected Xiongguo base URL: %s",got)
 }
}

func TestApk391XiongguoVerifiedHeaderCategories(t *testing.T) {
 cats:=apk391StaticCategoryFallback["xiongguo"]
 if len(cats)!=27{t.Fatalf("expected 3 sections and 24 video categories, got %d",len(cats))}
 want:=map[string]string{"1":"傳媒","2":"視頻","3":"電影","6":"91傳媒","13":"星空傳媒","14":"自拍偷拍","21":"onlyfans","22":"中文字幕","29":"H動畫"}
 got:=map[string]string{}
 for _,cat:=range cats {
  if !apk391ValidCategory("xiongguo",cat.ID){t.Fatalf("invalid verified category %s",cat.ID)}
  if _,exists:=got[cat.ID];exists{t.Fatalf("duplicate category ID %s",cat.ID)}
  got[cat.ID]=cat.Name
 }
 for id,name:=range want{if got[id]!=name{t.Fatalf("category %s name %q, expected %q",id,got[id],name)}}
 if apk391ValidCategory("xiongguo","290"){t.Fatal("unrelated advertising category should be rejected")}
}

func TestApk391MangguoDetailRetainsCatalogTitle(t *testing.T) {
 const source="mangguo"
 const id="987654321"
 const original="芒果片單已確認的正確片名"
 apk391RememberTitle(source,id,original)
 t.Cleanup(func(){apk391CatalogNames.Delete(source+":"+id)})
 d:=sourceFixtureDownloader(t,func(req *http.Request)(*http.Response,error){
  if req.URL.Path=="/index.php/vod/detail/id/"+id+".html" {
   return sourceFixtureResponse(req,http.StatusOK,`<html><body><h1>不相關的其他影片</h1><a href="/index.php/vod/play/id/987654321/sid/1/nid/1.html">第1集</a></body></html>`),nil
  }
  return sourceFixtureResponse(req,http.StatusNotFound,"missing"),nil
 })
 drama,_,err:=d.fetchApk391Detail(context.Background(),source,id)
 if err!=nil{t.Fatal(err)}
 if drama.Title!=original||drama.Name!=original{t.Fatalf("detail overwrote original catalog title: %+v",drama)}
}

func TestChengguoBrowseArticleLazyPosters(t *testing.T) {
 htmlText:=`<html><body><div data-xpch="card-list">
 <article data-xpch="card-drama"><div><img src="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///"></div><h3>第一部短劇</h3><a href="/drama/dj-a123" aria-label="第一部短劇 详情">详情</a><a href="/play/dj-a123/3"></a></article>
 <article data-xpch="card-drama"><div><img src="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///"></div><h3>第二部短劇</h3><a href="/drama/dj-b456" aria-label="第二部短劇 详情">详情</a><a href="/play/dj-b456/2"></a></article>
 <article data-xpch="card-drama"><div><img src="/_img/poster.jpeg"></div><h3>第三部短劇</h3><a href="/drama/dj-c789">详情</a></article>
 </div><a href="/browse/page-2">下一頁</a></body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(htmlText));if err!=nil{t.Fatal(err)}
 got:=chengguoCards(doc)
 if len(got)!=3{t.Fatalf("expected 3 movies despite lazy placeholders, got %d: %+v",len(got),got)}
 for i,id:=range []string{"dj-a123","dj-b456","dj-c789"}{
  if got[i].SourceID!=id{t.Fatalf("film %d ID mismatch: %s",i,got[i].SourceID)}
 }
 if got[0].Cover!=""||got[1].Cover!=""{t.Fatalf("transparent placeholders leaked as posters: %+v",got)}
 if !chengguoHasNext(doc,1){t.Fatal("missing verified browse/page-2 continuation")}
}

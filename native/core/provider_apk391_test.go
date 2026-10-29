package core

import (
 "context"
 "net/http"
 "strings"
 "testing"

 nethtml "golang.org/x/net/html"
)

func TestApk391NativeIdentityAndIsolation(t *testing.T) {
 if len(apk391Sources)!=45 || len(apk391ByID)!=45 {
  t.Fatalf("expected 45 unique sources, got %d / %d",len(apk391Sources),len(apk391ByID))
 }
 for _,spec:=range apk391Sources {
  if !isHuangguoProviderSource(spec.ID) || canonicalProviderSource(spec.ID)!=spec.ID{
   t.Fatalf("registered source not recognized: %+v",spec)
  }
  if apk391BaseURL(spec.ID)=="" || providerSourceForURL("https://"+spec.Host+"/")!=spec.ID{
   t.Fatalf("host/source mapping lost: %+v",spec)
  }
 }
 for _,id:=range []string{"lujj31","liuguo","xiaoguo","hongguo"}{
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
 if total<436 || groups<19 { t.Fatalf("missing APK static classifications: %d rows, %d groups",total,groups)}
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

func TestApk391UnverifiedCustomSourceHasExplicitError(t *testing.T){
 d:=sourceFixtureDownloader(t,func(r *http.Request)(*http.Response,error){
  t.Fatalf("should not call unknown custom API: %s",r.URL)
  return sourceFixtureResponse(r,http.StatusInternalServerError,""),nil
 })
 _,_,err:=d.fetchApk391CatalogPage(context.Background(),"xiongguo",1,"","")
 if err==nil||!strings.Contains(err.Error(),"尚未"){t.Fatalf("custom source should explain limitation, got %v",err)}
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
  <section class="module-item"><a href="/voddetail/901.html"><img alt="真正影片" data-src="/cover901.jpg"></a></section>
  </body></html>`
  return sourceFixtureResponse(r,http.StatusOK,body),nil
 })
 items,more,err:=d.fetchApk391CatalogPage(context.Background(),"yanguo",1,"","")
 if err!=nil||len(items)!=1||items[0].Title!="真正影片"||!more {
  t.Fatalf("expected real movie only: %+v more=%v err=%v",items,more,err)
 }
}

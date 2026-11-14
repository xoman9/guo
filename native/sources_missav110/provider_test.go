package missavrecovery

import (
 "context"
 "errors"
 "strings"
 "testing"

 "duanjuapp/native/core"
 nethtml "golang.org/x/net/html"
)

func missTestDoc(t *testing.T,html string)*nethtml.Node {
 t.Helper()
 doc,e:=nethtml.Parse(strings.NewReader(html));if e!=nil{t.Fatal(e)}
 return doc
}
func TestWebHomeMissCategories(t *testing.T){
 cats:=missCategories(nil)
 if len(cats)<30 || cats[1].ID!="cnsub" || cats[1].Name!="中文字幕"{
  t.Fatalf("WebHome categories: %+v",cats)
 }
 for _,id:=range []string{"cnsub","new","vr","fc2","madou","twav"}{
  p,ok:=webHomeCategoryPath(id)
  if !ok||p==""||!validCategory(id){t.Fatalf("missing WebHome route %q",id)}
 }
 if validCategory("../other")||validCategory("not-listed"){t.Fatal("unapproved category accepted")}
}
func TestJavaMissScopedThumbnailCards(t *testing.T){
 doc:=missTestDoc(t,`<div class="thumbnail"><a class="text-secondary" href="/abc-123">影片一</a><img data-src="//img.example.test/one.jpg"><span>更新</span></div>
 <div class="thumbnail"><a class="text-secondary" href="https://missav.ws/Ab-9">影片二</a><img src="/two.jpg"></div>
 <div class="thumbnail"><a class="text-secondary" href="/abc-123">重複</a></div>
 <div class="thumbnail"><a class="text-secondary" href="https://evil.test/x">外站</a><img src="/bad.jpg"></div>`)
 rows:=missCards(doc)
 if len(rows)!=2||rows[0].SourceID!="abc-123"||rows[1].SourceID!="Ab-9"||rows[0].Cover!="https://img.example.test/one.jpg"||rows[1].Cover!="https://missav.ws/two.jpg"{
  t.Fatalf("Java div.thumbnail parsing: %+v",rows)
 }
}
type missFixtureHost struct {
 core.Host
 pages map[string]string
 state map[string][]byte
 last core.TextRequest
}
func(f *missFixtureHost)SaveState(key string,data []byte)error{
 if f.state==nil{f.state=map[string][]byte{}}
 f.state[key]=append([]byte(nil),data...)
 return nil
}
func(f *missFixtureHost)LoadState(key string)([]byte,error){return f.state[key],nil}
func(f *missFixtureHost)FetchText(_ context.Context,req core.TextRequest)(core.TextResponse,error){
 f.last=req
 if body,ok:=f.pages[req.URL];ok{return core.TextResponse{Body:body,URL:req.URL},nil}
 return core.TextResponse{},errors.New("missing fixture "+req.URL)
}
func TestMissAVUsesAppLocalVerifiedSessionForCatalogAndDetail(t *testing.T) {
 host:=&missFixtureHost{pages:map[string]string{
  baseURL+"/":`<div class="thumbnail"><a class="text-secondary" href="/abc-123">卡片</a></div>`,
  baseURL+"/abc-123":`<meta property="og:title" content="片名">`,
 }}
 if err:=core.SaveVerifiedSiteSession(host,sourceID,core.VerifiedSiteSession{
  Cookie:"verified=session",UserAgent:"Android WebView",
 });err!=nil{t.Fatal(err)}
 p:=&provider{host:host}
 rows,_,err:=p.CatalogPage(context.Background(),1,"")
 if err!=nil||len(rows)!=1{t.Fatalf("catalog with session: %v %+v",err,rows)}
 if host.last.Cookie!="verified=session"||host.last.UserAgent!="Android WebView"{
  t.Fatalf("catalog did not use app-local verified session: %+v",host.last)
 }
 _,_,err=p.Detail(context.Background(),"abc-123")
 if err!=nil{t.Fatalf("detail with session: %v",err)}
 if host.last.Cookie!="verified=session"||host.last.URL!=baseURL+"/abc-123"{
  t.Fatalf("detail did not use app-local verified session: %+v",host.last)
 }
}

func TestJavaMissProviderRoutes(t *testing.T){
 host:=&missFixtureHost{pages:map[string]string{
  baseURL+"/":`<a class="block bg-nord3" href="/dm">麻豆</a><div class="thumbnail"><a class="text-secondary" href="/abc-123">首頁影片</a><img src="/a.jpg"></div>`,
  baseURL+"/dm539/cn/new?page=2":`<div class="thumbnail"><a class="text-secondary" href="/xyz-34">第二頁</a><img src="/b.jpg"></div><a rel="next" href="/dm539/cn/new?page=3">下一頁</a>`,
  baseURL+"/cn/search/hello":`<div class="thumbnail"><a class="text-secondary" href="/abc-123">搜尋結果</a><img src="/a.jpg"></div>`,
  baseURL+"/abc-123":`<html><head><meta property="og:title" content="影片一"><meta property="og:image" content="/poster.jpg"></head></html>`,
 }}
 p:=&provider{host:host}
 cats,e:=p.Categories(context.Background(),false)
 if e!=nil||len(cats)<30||cats[1].ID!="cnsub"{t.Fatalf("categories %+v %v",cats,e)}
 rows,more,e:=p.CatalogPage(context.Background(),2,"new")
 if e!=nil||len(rows)!=1||rows[0].SourceID!="xyz-34"||!more{t.Fatalf("catalog %+v more=%t %v",rows,more,e)}
 rows,_,e=p.Search(context.Background(),1,"hello")
 if e!=nil||len(rows)!=1{t.Fatalf("search %+v %v",rows,e)}
 drama,ch,e:=p.Detail(context.Background(),"abc-123")
 if e!=nil||drama.Title!="影片一"||drama.Cover!="https://missav.ws/poster.jpg"||len(ch)!=1{t.Fatalf("detail %+v %+v %v",drama,ch,e)}
 task:=core.Task{DramaID:drama.ID,Chapter:ch[0]}
 path,e:=mediaPageForTask(task)
 if e!=nil||path!="/abc-123"{t.Fatalf("media route %q %v",path,e)}
 task.Chapter.PageURL=baseURL+"/other"
 if _,e:=mediaPageForTask(task);e==nil{t.Fatal("accepted mismatched media page")}
}
func TestWebHomePreviewUUIDPlaysWithoutDetailHTTP(t *testing.T) {
 uuid:="12345678-1234-1234-1234-123456789abc"
 host:=&missFixtureHost{pages:map[string]string{
  baseURL+"/":`<div class="thumbnail">
    <a class="text-secondary" href="/abc-123">卡片直接播放</a>
    <img src="/cover.jpg">
    <div id="preview-12345678-1234-1234-1234-123456789abc"></div>
   </div>`,
 }}
 p:=&provider{host:host}
 rows,_,err:=p.CatalogPage(context.Background(),1,"")
 if err!=nil||len(rows)!=1{t.Fatalf("catalog: %+v %v",rows,err)}
 info,ok:=p.card("abc-123")
 if !ok||info.uuid!=uuid{t.Fatalf("UUID not saved: %+v %v",info,ok)}
 // The fixture intentionally provides no media detail page; this must work
 // with only the listing card, just like WebHome.
 drama,ch,err:=p.Detail(context.Background(),"abc-123")
 if err!=nil||len(ch)!=1||drama.Title!="卡片直接播放"{
  t.Fatalf("detail from cached card: %+v %+v %v",drama,ch,err)
 }
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:ch[0]})
 expected:="https://surrit.com/"+uuid+"/playlist.m3u8"
 if err!=nil||media.URL!=expected||media.Referer!=baseURL+"/"{
  t.Fatalf("media from WebHome preview: %+v %v",media,err)
 }
}
func TestWebHomePreviewUUIDNotBorrowedFromOtherCard(t *testing.T){
 doc:=missTestDoc(t,`<div class="thumbnail"><a class="text-secondary" href="/abc-123">No UUID</a>
 <img src="/cover.jpg"></div>
 <div class="thumbnail"><a class="text-secondary" href="/def-456">With UUID</a>
 <div id="preview-12345678-1234-1234-1234-123456789abc"></div></div>
 <div id="preview-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"></div>`)
 p:=&provider{}
 rows:=p.cacheCards(doc)
 if len(rows)!=2{t.Fatalf("cards: %+v",rows)}
 if _,ok:=p.card("abc-123");ok{t.Fatal("borrowed UUID from unrelated card")}
 if info,ok:=p.card("def-456");!ok||info.uuid!="12345678-1234-1234-1234-123456789abc"{
  t.Fatalf("wrong scoped UUID: %+v %t",info,ok)
 }
}
func TestWebHomePackedPlaybackScript(t *testing.T) {
 page:=baseURL+"/abc-123"
 expected:="https://surrit.com/12345678-1234-1234-1234-123456789abc/playlist.m3u8"
 // Minimal Dean Edwards packer fixture equivalent to WebHome unpackEvals.
 packed:=`<script>eval(function(p,a,c,k,e,d){}('0://1/2/playlist.m3u8',10,3,'https|surrit.com|12345678-1234-1234-1234-123456789abc'.split('|'),0,{}))</script>`
 got,err:=directMediaURL(page,packed)
 if err!=nil||got!=expected {t.Fatalf("packed WebHome media: %q, %v",got,err)}
 if out:=unpackMissPacker(`<script>var x='12345678-1234-1234-1234-123456789abc'</script>`);out!=""{
  t.Fatalf("unpacked nonpacker script: %q",out)
 }
}

func TestMissAVPrefersCanonicalMasterWhenPageHasMultipleMedia(t *testing.T) {
 page:=baseURL+"/abc-123"
 playlist:="https://surrit.com/12345678-1234-1234-1234-123456789abc/playlist.m3u8"
 html:=`<video src="https://cdn.example.test/preview.mp4"></video><script>var master="`+playlist+`"</script>`
 got,err:=directMediaURL(page,html)
 if err!=nil||got!=playlist{t.Fatalf("should prefer master HLS: %q %v",got,err)}
 html+=`<script>var other="https://surrit.com/abcdef12-1234-1234-1234-123456789abc/playlist.m3u8"</script>`
 if got,err:=directMediaURL(page,html);err==nil{
  t.Fatalf("ambiguous HLS masters accepted: %q",got)
 }
}

func TestMissAVSeekAdjacentStreamID(t *testing.T) {
 page:=baseURL+"/abc-123"
 uuid:="12345678-1234-1234-1234-123456789abc"
 expected:="https://surrit.com/"+uuid+"/playlist.m3u8"
 for _,body:=range []string{
  `<script>var stream="12345678-1234-1234-1234-123456789abc",seek=0;</script>`,
  `<script>player("12345678-1234-1234-1234-123456789abc");seek(0)</script>`,
 }{
  got,err:=directMediaURL(page,body)
  if err!=nil||got!=expected{t.Fatalf("seek-adjacent stream: %q %v",got,err)}
 }
 for _,body:=range []string{
  `<meta content="12345678-1234-1234-1234-123456789abc"><script>seek(0)</script>`,
  `<script>var unrelated="12345678-1234-1234-1234-123456789abc"; var text="1234567890123456789012345678901234567890"; seek(0);</script>`,
 }{
  if got,err:=directMediaURL(page,body);err==nil{t.Fatalf("guessed unrelated stream: %q",got)}
 }
}

func TestWebHomeInlineSurritPlaylist(t *testing.T){
 page:=baseURL+"/abc-123"
 playlist:="https://surrit.com/12345678-1234-1234-1234-123456789abc/playlist.m3u8"
 cases:=[]string{
  `<script>const stream = "`+playlist+`";</script>`,
  `<script>const stream = "https:\/\/surrit.com\/12345678-1234-1234-1234-123456789abc\/playlist.m3u8";</script>`,
 }
 for _,body:=range cases{
  got,err:=directMediaURL(page,body)
  if err!=nil||got!=playlist{t.Fatalf("WebHome HLS: %q %v",got,err)}
 }
 for _,body:=range []string{
  `<script>const id="12345678-1234-1234-1234-123456789abc";</script>`,
  `<script>const stream="https://other.example/12345678-1234-1234-1234-123456789abc/playlist.m3u8";</script>`,
 }{
  if got,err:=directMediaURL(page,body);err==nil{t.Fatalf("guessed media from unsupported script: %q",got)}
 }
}

func TestMissAVDirectMediaOnly(t *testing.T){
 page:=baseURL+"/abc-123"
 for _,body:=range []string{`<video src="https://cdn.example.test/clip.mp4"></video>`,`{"src":"https:\/\/cdn.example.test\/clip.mp4"}`}{
  u,e:=directMediaURL(page,body)
  if e!=nil||u==""{t.Fatalf("direct media rejected: %v",e)}
 }
 for _,body:=range []string{`<script>player.loadDynamic()</script>`,`<source src="javascript:alert(1)">`,`<video src="https://a.example/one.mp4"></video><video src="https://b.example/two.mp4"></video>`}{
  if u,e:=directMediaURL(page,body);e==nil{t.Fatalf("unsupported parser accepted %q",u)}
 }
}

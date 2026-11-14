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
}
func(f *missFixtureHost)FetchText(_ context.Context,req core.TextRequest)(core.TextResponse,error){
 if body,ok:=f.pages[req.URL];ok{return core.TextResponse{Body:body,URL:req.URL},nil}
 return core.TextResponse{},errors.New("missing fixture "+req.URL)
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

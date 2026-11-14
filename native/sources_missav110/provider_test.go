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
func TestJavaMissCategories(t *testing.T){
 doc:=missTestDoc(t,`<a class="block px-4 py-2 text-sm leading-5 text-nord5 bg-nord3" href="https://missav.com/dm">麻豆</a>
 <a class="block bg-nord3" href="/VR-works">VR</a>
 <a class="block bg-nord3" href="/dm">重複</a><a class="block bg-nord3" href="https://other.test/dm">不允許</a>`)
 cats:=missCategories(doc)
 if len(cats)!=3||cats[1].ID!="dm"||cats[1].Name!="麻豆"||cats[2].ID!="VR-works"{
  t.Fatalf("Java homeContent categories: %+v",cats)
 }
}
func TestJavaMissScopedThumbnailCards(t *testing.T){
 doc:=missTestDoc(t,`<div class="thumbnail"><a class="text-secondary" href="/abc-123">影片一</a><img data-src="//img.example.test/one.jpg"><span>更新</span></div>
 <div class="thumbnail"><a class="text-secondary" href="https://missav.com/Ab-9">影片二</a><img src="/two.jpg"></div>
 <div class="thumbnail"><a class="text-secondary" href="/abc-123">重複</a></div>
 <div class="thumbnail"><a class="text-secondary" href="https://evil.test/x">外站</a><img src="/bad.jpg"></div>`)
 rows:=missCards(doc)
 if len(rows)!=2||rows[0].SourceID!="abc-123"||rows[1].SourceID!="Ab-9"||rows[0].Cover!="https://img.example.test/one.jpg"||rows[1].Cover!="https://missav.com/two.jpg"{
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
  baseURL+"/dm?page=2":`<div class="thumbnail"><a class="text-secondary" href="/xyz-34">第二頁</a><img src="/b.jpg"></div><a rel="next" href="/dm?page=3">下一頁</a>`,
  baseURL+"/search/hello?page=1":`<div class="thumbnail"><a class="text-secondary" href="/abc-123">搜尋結果</a><img src="/a.jpg"></div>`,
  baseURL+"/abc-123":`<html><head><meta property="og:title" content="影片一"><meta property="og:image" content="/poster.jpg"></head></html>`,
 }}
 p:=&provider{host:host}
 cats,e:=p.Categories(context.Background(),false)
 if e!=nil||len(cats)!=2||cats[1].ID!="dm"{t.Fatalf("categories %+v %v",cats,e)}
 rows,more,e:=p.CatalogPage(context.Background(),2,"dm")
 if e!=nil||len(rows)!=1||rows[0].SourceID!="xyz-34"||!more{t.Fatalf("catalog %+v more=%t %v",rows,more,e)}
 rows,_,e=p.Search(context.Background(),1,"hello")
 if e!=nil||len(rows)!=1{t.Fatalf("search %+v %v",rows,e)}
 drama,ch,e:=p.Detail(context.Background(),"abc-123")
 if e!=nil||drama.Title!="影片一"||drama.Cover!="https://missav.com/poster.jpg"||len(ch)!=1{t.Fatalf("detail %+v %+v %v",drama,ch,e)}
 task:=core.Task{DramaID:drama.ID,Chapter:ch[0]}
 path,e:=mediaPageForTask(task)
 if e!=nil||path!="/abc-123"{t.Fatalf("media route %q %v",path,e)}
 task.Chapter.PageURL=baseURL+"/other"
 if _,e:=mediaPageForTask(task);e==nil{t.Fatal("accepted mismatched media page")}
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

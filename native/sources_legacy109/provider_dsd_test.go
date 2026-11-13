package legacy109
import (
 "context"
 "errors"
 "strings"
 "testing"
 "duanjuapp/native/core"
)
type dsd109Fixture struct{core.Host;routes map[string]string}
func (f *dsd109Fixture) FetchText(_ context.Context,req core.TextRequest)(core.TextResponse,error){
 body,ok:=f.routes[req.URL];if !ok{return core.TextResponse{},errors.New("missing fixture: "+req.URL)}
 return core.TextResponse{Body:body},nil
}
func TestDSD109CategoriesCatalogDetailAndPlayback(t *testing.T){
 root:=`<a href="/index.php/vod/type/id/3.html">電影</a>`
 cards:=`<a class="video-item" href="/index.php/vod/detail/id/100.html" title="樣本"><img data-src="/a.png"></a>`
 detail:=`<h1>樣本</h1><a href="/index.php/vod/play/id/100/sid/1/nid/2.html">第2集</a><script>var player_aaaa={"id":"100","sid":1,"nid":1,"url":"https://media.example/ep1.mp4","encrypt":0};</script>`
 play:=`<script>var player_aaaa={"id":"100","sid":1,"nid":2,"url":"https://media.example/ep2.mp4","encrypt":0};</script>`
 h:=&dsd109Fixture{routes:map[string]string{
  dsd109Base+"/":root,
  dsd109Base+"/index.php/vod/type/id/3/page/1.html":cards,
  dsd109Base+dsd109Page("100",1,1):detail,
  dsd109Base+dsd109Page("100",1,2):play,
 }}
 p:=&dsd109{host:h}
 cats,err:=p.Categories(context.Background(),false)
 if err!=nil||len(cats)!=1||cats[0].ID!="3"{t.Fatalf("cats %+v %v",cats,err)}
 rows,_,err:=p.CatalogPage(context.Background(),1,"3")
 if err!=nil||len(rows)!=1||rows[0].SourceID!="100"{t.Fatalf("catalog %+v %v",rows,err)}
 drama,chapters,err:=p.Detail(context.Background(),"100")
 if err!=nil||drama.Title!="樣本"||len(chapters)!=2{t.Fatalf("detail %+v %+v %v",drama,chapters,err)}
 chosen:=chapters[0];for _,c:=range chapters{if strings.HasSuffix(c.ID,"1-2"){chosen=c}}
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chosen})
 if err!=nil||media.URL!="https://media.example/ep2.mp4"{t.Fatalf("media %+v %v",media,err)}
}
func TestDSD109IdentityRejectsWrongEpisodeAndHost(t *testing.T){
 if _,_,ok:=dsd109Route(dsd109Base+"/","https://evil.example/index.php/vod/type/id/3.html");ok{t.Fatal("external URL accepted")}
 h:=&dsd109Fixture{routes:map[string]string{
 dsd109Base+dsd109Page("100",1,2):`<script>var player_aaaa={"id":"100","sid":1,"nid":1,"url":"https://example.com/preview.mp4"};</script>`,
 }}
 _,err:=(&dsd109{host:h}).ResolveMedia(context.Background(),core.Task{
 DramaID:core.ProviderDramaID("dsd","100"),
 Chapter:core.Chapter{ID:core.ProviderChapterID("dsd","100","1-2")},
 })
 if err==nil{t.Fatal("wrong episode accepted")}
}

package legacy109

import (
 "context"
 "errors"
 "strings"
 "testing"

 "duanjuapp/native/core"
)

func Test109HeitaoCatalogRoutes(t *testing.T) {
 cases:=[]struct{page int; category,search,want string}{
  {1,"","","/index.php/vod/show.html"},
  {2,"3","","/index.php/vod/show/id/3/page/2.html"},
  {1,"","movie","/index.php/vod/search/wd/movie.html"},
 }
 for _,c:=range cases {
  got,err:=heitaoCatalogRoute(c.page,c.category,c.search)
  if err!=nil||got!=c.want{t.Fatalf("route %q want %q err %v",got,c.want,err)}
 }
 if _,err:=heitaoCatalogRoute(2,"../../","");err==nil{t.Fatal("invalid category accepted")}
}
func Test109HeitaoCardsAndEpisodes(t *testing.T) {
 cards:=`<a class="thumb" href="/index.php/vod/detail/id/91.html" title="劇集 91"><img data-src="/cover.jpg"></a>`
 rows:=heitaoParseCards(cards)
 if len(rows)!=1||rows[0].SourceID!="91"||rows[0].Title!="劇集 91"{t.Fatalf("unexpected cards: %+v",rows)}
 ep:=`<a href="/index.php/vod/play/id/91/sid/2/nid/1.html">第一集</a><a href="/index.php/vod/play/id/91/sid/2/nid/2.html">第二集</a>`
 chapters:=heitaoParseEpisodes(ep,"91")
 if len(chapters)!=2||chapters[0].EpisodeString(0)!="1"||chapters[1].EpisodeString(0)!="2"{t.Fatalf("bad episodes: %+v",chapters)}
}

func Test109HeitaoNestedPlayerJSON(t *testing.T) {
 raw:=`<script>var player_aaaa={"url":"https://cdn.example/movie.m3u8","encrypt":0,"meta":{"nested":{},"quoted":"}"}};</script>`
 entry,err:=heitaoReadPlayerConfig109(raw)
 if err!=nil||entry.URL!="https://cdn.example/movie.m3u8"{t.Fatalf("nested JSON failed: %+v %v",entry,err)}
 if _,err:=heitaoReadPlayerConfig109(`player_aaaa={"url":"oops"`);err==nil{t.Fatal("incomplete JSON accepted")}
}

type heitaoFixtureHost struct{
 core.Host
 pages map[string]string
 calls int
}
func (h *heitaoFixtureHost) FetchText(_ context.Context,r core.TextRequest)(core.TextResponse,error){
 h.calls++
 if body,ok:=h.pages[r.URL];ok{return core.TextResponse{Body:body,URL:r.URL},nil}
 return core.TextResponse{},errors.New("unmocked black-peach page: "+r.URL)
}
func Test109HeitaoPlayerRouteBindsSeriesAndEpisode(t *testing.T){
 base:=core.Task{
  DramaID:core.ProviderDramaID("heitao","91"),
  Chapter:core.Chapter{
   ID:core.ProviderChapterID("heitao","91","2:3"),
   PageURL:htBase+"/index.php/vod/player/id/91/nid/3/sid/2.html",
  },
 }
 route,err:=heitaoPlayerRoute(base)
 if err!=nil||route!="/index.php/vod/player/id/91/nid/3/sid/2.html"{
  t.Fatalf("valid player route rejected: %q %v",route,err)
 }
 cases:=[]core.Task{}
 wrong:=base
 wrong.DramaID=core.ProviderDramaID("heitao","92")
 cases=append(cases,wrong)
 wrong=base
 wrong.Chapter.ID=core.ProviderChapterID("heitao","91","2:4")
 cases=append(cases,wrong)
 wrong=base
 wrong.Chapter.PageURL=htBase+"/index.php/vod/player/id/92/nid/3/sid/2.html"
 cases=append(cases,wrong)
 wrong=base
 wrong.Chapter.PageURL=htBase+".evil.test/index.php/vod/player/id/91/nid/3/sid/2.html"
 cases=append(cases,wrong)
 wrong=base
 wrong.Chapter.ID=core.ProviderChapterID("heitao","91","../../")
 cases=append(cases,wrong)
 wrong=base
 wrong.Chapter.PageURL=base.Chapter.PageURL+"?redirect=1"
 cases=append(cases,wrong)
 for _,task:=range cases{
  if _,err:=heitaoPlayerRoute(task);err==nil{t.Fatalf("mismatch accepted: %+v",task)}
 }
 host:=&heitaoFixtureHost{pages:map[string]string{
  htBase+route:`<script>var player_aaaa={"url":"https://cdn.example.net/episode3.mp4","encrypt":0};</script>`,
 }}
 p:=&heitaoProvider{host:host}
 media,err:=p.ResolveMedia(context.Background(),base)
 if err!=nil||media.URL!="https://cdn.example.net/episode3.mp4"{
  t.Fatalf("valid heitao episode failed: %+v %v",media,err)
 }
 before:=host.calls
 wrong=base
 wrong.Chapter.ID=core.ProviderChapterID("heitao","91","2:4")
 if _,err:=p.ResolveMedia(context.Background(),wrong);err==nil||!strings.Contains(err.Error(),"does not match"){
  t.Fatalf("wrong heitao episode accepted: %v",err)
 }
 if host.calls!=before{t.Fatal("mismatched chapter triggered a network request")}
}

func Test109HeitaoEqualLengthPlaylistsChooseStableSID(t *testing.T){
 html:=`<a href="/index.php/vod/play/id/91/sid/3/nid/1.html">線路三</a>
<a href="/index.php/vod/play/id/91/sid/2/nid/1.html">線路二</a>`
 for i:=0;i<200;i++{
  chapters:=heitaoParseEpisodes(html,"91")
  if len(chapters)!=1||chapters[0].PageURL!=htBase+"/index.php/vod/player/id/91/nid/1/sid/2.html"{
   t.Fatalf("unstable playlist on iteration %d: %+v",i,chapters)
  }
 }
}

func Test109HeitaoCatalogLinksRejectForeignHost(t *testing.T){
 cases:=[]struct{link,want string}{
  {"/index.php/vod/detail/id/91.html","91"},
  {"https://heitaodj.com/index.php/vod/detail/id/91.html","91"},
  {"https://foreign.example/index.php/vod/detail/id/91.html",""},
  {"https://heitaodj.com.evil.test/index.php/vod/detail/id/91.html",""},
  {"//evil.test/index.php/vod/detail/id/91.html",""},
  {"/index.php/vod/detail/id/91.html/more",""},
  {"/index.php/vod/detail/id/%2e%2e.html",""},
  {"/index.php/vod/detail/id/91.html?sort=hot","91"},
 }
 for _,tc:=range cases{
  if got:=heitaoCardID(tc.link);got!=tc.want{t.Errorf("link %q: got %q want %q",tc.link,got,tc.want)}
 }
 page:=`<a class="thumb" href="https://bad.example/index.php/vod/detail/id/101.html" title="Bad"><img src="/bad.jpg"></a>
<a class="thumb" href="/index.php/vod/detail/id/102.html" title="Good"><img src="/good.jpg"></a>`
 cards:=heitaoParseCards(page)
 if len(cards)!=1||cards[0].SourceID!="102"{
  t.Fatalf("unexpected external card: %+v",cards)
 }
}

package legacy109

import "testing"

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

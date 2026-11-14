package legacy109

import (
 "context"
 "errors"
 "strings"
 "testing"

 "duanjuapp/native/core"
)

type hongguoFixtureHost struct{
 core.Host
 pages map[string]string
}
func (h *hongguoFixtureHost) FetchText(_ context.Context,req core.TextRequest)(core.TextResponse,error){
 if body,ok:=h.pages[req.URL];ok{return core.TextResponse{Body:body,URL:req.URL},nil}
 return core.TextResponse{},errors.New("unmocked Hongguo route: "+req.URL)
}
func hgFixture(body string)string{
 return "<script>window._ROUTER_DATA = "+body+";</script>"
}
func TestHongguo109CategoriesAndCatalog(t *testing.T){
 host:=&hongguoFixtureHost{pages:map[string]string{
  hgBase+"/category/real-drama?page=1":hgFixture(
   "{\"loaderData\":{\"category_page\":{\"isSuccess\":true,\"recommendList\":[{\"video_data\":{\"series_id_str\":\"211\",\"series_title\":\"範例短劇\",\"series_cover\":\"/cover.jpg\",\"episode_cnt\":\"2\"}}],\"pagination\":{\"totalPages\":2}}}}"),
 }}
 p:=&hongguo109{host:host}
 categories,err:=p.Categories(context.Background(),false)
 if err!=nil||len(categories)!=4{t.Fatalf("categories: %+v %v",categories,err)}
 items,more,err:=p.CatalogPage(context.Background(),1,"real-drama")
 if err!=nil||!more||len(items)!=1||items[0].SourceID!="211"||items[0].Title!="範例短劇"{
  t.Fatalf("catalog: %+v %t %v",items,more,err)
 }
 if _,_,err=p.CatalogPage(context.Background(),1,"../bad");err==nil{t.Fatal("invalid category accepted")}
}
func TestHongguo109DetailAndNoPreviewSubstitution(t *testing.T){
 ctx:=context.Background()
 host:=&hongguoFixtureHost{pages:map[string]string{
  hgBase+"/detail?series_id=211":hgFixture(
   "{\"loaderData\":{\"detail_page\":{\"seriesDetail\":{\"series_name\":\"範例短劇\",\"vid_list\":[\"31\",\"32\"]}}}}"),
  hgBase+"/player/211/31":hgFixture(
   "{\"loaderData\":{\"player_page\":{\"vid\":\"31\",\"series_id\":\"211\",\"video_player_info\":{\"url\":\"https://cdn.example.org/e31.mp4\"}}}}"),
  hgBase+"/player/211/32":hgFixture(
   "{\"loaderData\":{\"player_page\":{\"vid\":\"999\",\"series_id\":\"211\",\"video_player_info\":{\"url\":\"https://cdn.example.org/preview.mp4\"}}}}"),
 }}
 p:=&hongguo109{host:host}
 drama,chapters,err:=p.Detail(ctx,"211")
 if err!=nil||drama.SourceID!="211"||len(chapters)!=2{t.Fatalf("detail: %+v %+v %v",drama,chapters,err)}
 media,err:=p.ResolveMedia(ctx,core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err!=nil||media.URL!="https://cdn.example.org/e31.mp4"{t.Fatalf("media: %+v %v",media,err)}
 _,err=p.ResolveMedia(ctx,core.Task{DramaID:drama.ID,Chapter:chapters[1]})
 if err==nil||!strings.Contains(err.Error(),"mismatch"){t.Fatalf("wrong preview not rejected: %v",err)}
}
func TestHongguo109DRMNotBypassed(t *testing.T){
 host:=&hongguoFixtureHost{pages:map[string]string{
  hgBase+"/player/211/31":hgFixture(
   "{\"loaderData\":{\"player_page\":{\"vid\":\"31\",\"series_id\":\"211\",\"video_player_info\":{\"cenc\":true,\"url\":\"https://cdn.example.org/e31.mp4\"}}}}"),
 }}
 p:=&hongguo109{host:host}
 _,err:=p.ResolveMedia(context.Background(),core.Task{
  DramaID:core.ProviderDramaID("hongguo","211"),
  Chapter:core.Chapter{ID:core.ProviderChapterID("hongguo","211","31")},
 })
 if err==nil||!strings.Contains(err.Error(),"authorization"){t.Fatalf("encrypted player wrongly accepted: %v",err)}
}

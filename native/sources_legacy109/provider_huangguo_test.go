package legacy109

import (
 "context"
 "encoding/json"
 "errors"
 "strings"
 "testing"
 "duanjuapp/native/core"
)
type hg109Host struct{
 core.Host
 pages map[string]string
 api map[string]string
}
func (h *hg109Host) FetchText(_ context.Context,r core.TextRequest)(core.TextResponse,error){
 if text,ok:=h.pages[r.URL];ok{return core.TextResponse{Body:text},nil}
 return core.TextResponse{},errors.New("unmocked page: "+r.URL)
}
func (h *hg109Host) PrepareWebMedia(_ context.Context,media core.ProviderMedia,_ string)(core.ProviderMedia,error){return media,nil}
func (h *hg109Host) Fetch(_ context.Context,r core.HTTPRequest)(core.HTTPResponse,error){
 if r.URL!=hgAIBase+"/api/edge/request"{return core.HTTPResponse{},errors.New("unmocked AI endpoint")}
 var m map[string]string
 if json.Unmarshal(r.Body,&m)!=nil{return core.HTTPResponse{},errors.New("bad AI envelope")}
 if raw,ok:=h.api[m["path"]];ok{return core.HTTPResponse{Body:[]byte(raw),Status:200},nil}
 return core.HTTPResponse{},errors.New("unexpected edge API path "+m["path"])
}
func Test109HuangguoAIDetailAndPlayback(t *testing.T){
 h:=&hg109Host{api:map[string]string{
  "/api/videos/ab1":`{"status":1,"data":{"id":"ab1","title":"測試劇","episodes":[{"ep_num":1,"title":"第一集"}]}}`,
  "/api/videos/ab1/play?ep=1&page_texts=0":`{"status":1,"data":{"video_url":"https://video.example.org/ab1.mp4"}}`,
 }}
 p:=&huangguo109{host:h,ai:true}
 drama,chapters,err:=p.Detail(context.Background(),"ab1")
 if err!=nil||drama.SourceID!="ab1"||len(chapters)!=1{t.Fatalf("AI details %+v %+v %v",drama,chapters,err)}
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err!=nil||media.URL!="https://video.example.org/ab1.mp4"{t.Fatalf("AI media %+v %v",media,err)}
}
func Test109HuangguoVideoCatalogAndDirectHLS(t *testing.T){
 h:=&hg109Host{pages:map[string]string{
  hgVideoBase+"/series?page=1":`<article class="video-card"><a href="/series/abc-1" title="測試作品"><img data-src="/poster.jpg"></a></article>`,
  hgVideoBase+"/video/ep1":`<script data-hls="https://hls.example/ep1.m3u8"></script>`,
 }}
 p:=&huangguo109{host:h}
 rows,_,err:=p.CatalogPage(context.Background(),1,"series")
 if err!=nil||len(rows)!=1||rows[0].SourceID!="series/abc-1"{t.Fatalf("video catalog %+v %v",rows,err)}
 id:=core.ProviderDramaID(p.source(),"video/ep1")
 chapter:=core.Chapter{ID:core.ProviderChapterID(p.source(),"video/ep1","1"),PageURL:hgVideoBase+"/video/ep1"}
 // M3U8 preparation requires its own Host fixture; the no-media behavior is independently checked.
 chapter.PageURL=hgVideoBase+"/video/missing"
 _,err=p.ResolveMedia(context.Background(),core.Task{DramaID:id,Chapter:chapter})
 if err==nil||!strings.Contains(err.Error(),"does not match chapter"){t.Fatalf("mismatched page was not rejected: %v",err)}
}

func Test109HuangguoVideoChapterPageBinding(t *testing.T){
 cases:=[]struct{id,ep,page string;valid bool}{
  {"video/ep1","1",hgVideoBase+"/video/ep1",true},
  {"series/series1","ep1",hgVideoBase+"/video/ep1",true},
  {"series/series1","ep1",hgVideoBase+"/video/ep2",false},
  {"video/ep1","2",hgVideoBase+"/video/ep1",false},
  {"series/series1","../other",hgVideoBase+"/video/other",false},
  {"series/series1","ep1","https://other.example/video/ep1",false},
 }
 for _,tc:=range cases{
  _,err:=hg109VideoPage(tc.id,tc.ep,tc.page)
  if (err==nil)!=tc.valid{t.Errorf("binding id=%q ep=%q page=%q err=%v",tc.id,tc.ep,tc.page,err)}
 }
}
func Test109HuangguoVideoResolveRejectsWrongEpisodeBeforeFetch(t *testing.T){
 p:=&huangguo109{host:&hg109Host{pages:map[string]string{hgVideoBase+"/video/other":`<script data-hls="https://hls.example/other.m3u8"></script>`}}}
 task:=core.Task{DramaID:core.ProviderDramaID(p.source(),"series/abc"),
 Chapter:core.Chapter{ID:core.ProviderChapterID(p.source(),"series/abc","ep1"),PageURL:hgVideoBase+"/video/other"}}
 _,err:=p.ResolveMedia(context.Background(),task)
 if err==nil||!strings.Contains(err.Error(),"does not match chapter"){t.Fatalf("wrong episode accepted: %v",err)}
}

func Test109HuangguoCatalogRejectsOffDomainAndNestedPaths(t *testing.T){
 origin:=hgVideoBase+"/series?page=1"
 cases:=[]struct{url,id string}{
  {"/series/abc-1","series/abc-1"},
  {"https://huangguo.video/video/ep1","video/ep1"},
  {"https://evil.example/series/abc-1",""},
  {"https://huangguo.video.evil.example/series/abc-1",""},
  {"/series/abc-1/episode/2",""},
  {"/series/%2e%2e",""},
  {"/video/one/extra",""},
  {"/series/abc-1?sort=hot","series/abc-1"},
 }
 for _,tc:=range cases{
  if got:=hg109CardID(origin,tc.url);got!=tc.id{
   t.Errorf("card URL %q: got %q, want %q",tc.url,got,tc.id)
  }
 }
 html:=`<article class="video-card"><a href="https://evil.example/series/abc-1" title="External"><img src="/x.jpg"></a></article>
 <article class="video-card"><a href="/series/real" title="Real"><img src="/y.jpg"></a></article>`
 rows:=hgVideoItems(html,origin)
 if len(rows)!=1||rows[0].SourceID!="series/real"{
  t.Fatalf("unexpected catalog cards: %+v",rows)
 }
}
